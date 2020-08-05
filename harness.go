package main

import (
	"debug/elf"
	"fmt"
	"log"
	"os"
	"runtime"
	"syscall"
	"time"
)

/*
 * Traps on every program syscall entry/exit and records the state.
 * Function should be called after the child process has been started
 * with ptrace enabled and has had syscall.Wait4 called on it once.
 * Continues program until exit or crash: ws will be updated with the status.
 * Takes as arguments the pid of process to trace and WaitStatus to update.
 * Also requires an argument identifying if target binary is an 64bit ELF
 * Returns an execTrace struct identifying the execution run.
 */
func traceSyscalls(pid int, ws *syscall.WaitStatus, is64bit bool) (execTrace, error) {
	var err error
	var regs syscall.PtraceRegs
	var curExecTrace execTrace

	// Set up timeout condition for crude infinite loop detection.
	tStart := time.Now()
	tEnd := tStart.Add(30 * time.Millisecond)

	for {
		// Timeout condition.
		tCur := time.Now()
		if tCur.After(tEnd) {
			return curExecTrace, nil
		}

		err = syscall.PtraceSyscall(pid, 0)
		if err != nil {
			return curExecTrace,
				fmt.Errorf("traceSyscalls failed to call PtraceSyscall: %s",
					err.Error())
		}

		_, err = syscall.Wait4(pid, ws, syscall.WALL, nil)
		if err != nil {
			return curExecTrace,
				fmt.Errorf("traceSyscalls failed to call Wait4: %s",
					err.Error())
		}

		// Return on program exit, crash or abort.
		if ws.Exited() == true ||
			ws.StopSignal() == syscall.SIGSEGV ||
			ws.StopSignal() == syscall.SIGABRT {
			return curExecTrace, nil
		}

		// Collect trace information.
		err = syscall.PtraceGetRegs(pid, &regs)
		if err != nil {
			return curExecTrace,
				fmt.Errorf("traceSyscalls failed to call PtraceGetRegs: %s",
					err.Error())
		}

		traceRegs := getInterestingRegs(&regs)
		curExecTrace.trace = append(curExecTrace.trace, traceRegs)

		// Also return on exit/exit_group syscalls
		// These sycall numbers depend on architecture.
		if is64bit {
			if regs.Orig_rax == 0x3c || regs.Orig_rax == 0xe7 {
				return curExecTrace, nil
			}
		} else {
			if regs.Orig_rax == 0x1 || regs.Orig_rax == 0xfc {
				return curExecTrace, nil
			}
		}
	}
}

/*
 * Runs a new harness instance as a new goroutine.
 * Should be called by a harness routine attempting to 'try again'
 * after hitting an error.
 * The calling harness is responsible for cleaning up resources and returning
 * after calling this function.
 */
func resetHarness(id int, cmd string,
	inputCases <-chan TestCase,
	interestCases chan<- TestCase,
	crashCases chan<- TestCase) {

	log.Printf("Harness %d encountered an error, resetting\n",
		id)
	go harness(id, cmd, inputCases, interestCases, crashCases)
}

/*
 * Harness will run the external binary specified by cmd and feed
 * it inputs from the inputCases channel. Interesting TestCases will
 * be placed in the interestCases output channel.
 */
func harness(id int, cmd string,
	inputCases <-chan TestCase,
	interestCases chan<- TestCase,
	crashCases chan<- TestCase) {

	var err error

	// Figure out architecture- only 32/64bit x86 binaries are supported.
	var is64bit = false
	elfFile, err := elf.Open(cmd)
	if err != nil {
		log.Fatalf("Failed to open %s: %s\n", cmd, err.Error())
	}

	if elfFile.FileHeader.Class == elf.ELFCLASS64 {
		is64bit = true
	}

	elfFile.Close()
	// List of unique execution traces for this harness.
	var uniqueTraces []execTrace

	var pAttr syscall.ProcAttr
	pAttr.Sys = &syscall.SysProcAttr{Ptrace: true}

	// Pipe needs to be created and given to new process to create
	// entry in /proc/pid/fd/, but otherwise is unused.
	procPipe, harnessPipe, err := os.Pipe()
	if err != nil {
		log.Printf("Harness with id %d failed to create pipe: %s\n",
			id, err.Error())
		resetHarness(id, cmd, inputCases, interestCases, crashCases)
		return
	}
	harnessPipe.Close()
	defer procPipe.Close()
	pAttr.Files = make([]uintptr, 1)
	pAttr.Files[0] = procPipe.Fd()

	// Lock OS thread as per syscall.SysProcAttr documentation.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	procPid, err := syscall.ForkExec(cmd, nil, &pAttr)
	if err != nil {
		log.Printf("Harness with id %d failed to start program: %s\n",
			id, err.Error())
		resetHarness(id, cmd, inputCases, interestCases, crashCases)
		return
	}

	// Child process recieves signal on startup due to ptrace.
	var ws syscall.WaitStatus
	_, err = syscall.Wait4(procPid, &ws, syscall.WALL, nil)
	if err != nil {
		log.Printf("Harness with id %d failed to wait for tracee: %s\n",
			id, err.Error())
		resetHarness(id, cmd, inputCases, interestCases, crashCases)
		return
	}

	// Run process until it's in a suitable state for snapshotting.
	err = setupSnapshotState(procPid, &ws, is64bit)
	if err != nil {
		log.Printf("Harness with id %d failed to set up snapshot state: %s\n",
			id, err.Error())
		resetHarness(id, cmd, inputCases, interestCases, crashCases)
		return
	}

	// Save process state for future restorations.
	var procSnapshot Snapshot
	procSnapshot, err = makeSnapshot(procPid)
	if err != nil {
		log.Printf("Harness with id %d failed to take a snapshot: %s\n",
			id, err.Error())
		resetHarness(id, cmd, inputCases, interestCases, crashCases)
		return

	}

	for inputCase := range inputCases {

		err = writeToProc(procPid, inputCase.input)
		if err != nil {
			log.Printf("Harness with id %d failed to write to process:"+
				"%s\n", id, err.Error())
			resetHarness(id, cmd, inputCases, interestCases, crashCases)
			return
		}

		// Trace execution and report back interesting cases.
		curExecTrace, err := traceSyscalls(procPid, &ws, is64bit)
		if err != nil {
			log.Printf("Harness with id %d failed to trace process:"+
				"%s\n", id, err.Error())
			resetHarness(id, cmd, inputCases, interestCases, crashCases)
			return
		}
		if isUniqueTrace(curExecTrace, uniqueTraces) {
			uniqueTraces = append(uniqueTraces, curExecTrace)
			// If the channel is full, we ignore the interesting case
			select {
			case interestCases <- inputCase:
			default:
			}

		}

		// Check for abort signals so we can ignore them.
		var aborted bool = false
		if ws.StopSignal() == syscall.SIGABRT {
			aborted = true
		}

		// Report segfaults and ignore other exit causes.
		if ws.StopSignal() == syscall.SIGSEGV && aborted == false {
			log.Printf("Harness with id %d crashed process with pid %d\n",
				id, procPid)
			crashCases <- inputCase
		}

		// Restore process state for next run.
		restoreSnapshot(procSnapshot)
	}
}
