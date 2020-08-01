package main

import (
	"debug/elf"
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
func traceSyscalls(pid int, ws *syscall.WaitStatus, is64bit bool) execTrace {
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
			return curExecTrace
		}

		err = syscall.PtraceSyscall(pid, 0)
		if err != nil {
			log.Fatal("traceSyscalls failed to call PtraceSyscall")
		}

		_, err = syscall.Wait4(pid, ws, syscall.WALL, nil)
		if err != nil {
			log.Fatal("traceSyscalls failed to call Wait4")
		}

		// Return on program exit, crash or abort.
		if ws.Exited() == true ||
			ws.StopSignal() == syscall.SIGSEGV ||
			ws.StopSignal() == syscall.SIGABRT {
			return curExecTrace
		}

		// Collect trace information.
		err = syscall.PtraceGetRegs(pid, &regs)
		if err != nil {
			log.Fatal("traceSyscalls failed to call PtraceGetRegs")
		}

		traceRegs := getInterestingRegs(&regs)
		curExecTrace.trace = append(curExecTrace.trace, traceRegs)

		// Also return on exit/exit_group syscalls
		// These sycall numbers depend on architecture.
		if is64bit {
			if regs.Orig_rax == 0x3c || regs.Orig_rax == 0xe7 {
				return curExecTrace
			}
		} else {
			if regs.Orig_rax == 0x1 || regs.Orig_rax == 0xfc {
				return curExecTrace
			}
		}
	}
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
		log.Fatalf("Harness with id %d failed to create pipe: %s\n",
			id, err.Error())
	}
	harnessPipe.Close()
	pAttr.Files = make([]uintptr, 1)
	pAttr.Files[0] = procPipe.Fd()

	// Lock OS thread as per syscall.SysProcAttr documentation.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	procPid, err := syscall.ForkExec(cmd, nil, &pAttr)
	if err != nil {
		log.Fatalf("Harness with id %d failed to start program: %s\n",
			id, err.Error())
	}

	// Child process recieves signal on startup due to ptrace.
	var ws syscall.WaitStatus
	_, err = syscall.Wait4(procPid, &ws, syscall.WALL, nil)
	if err != nil {
		log.Fatalf("Harness with id %d failed to wait for tracee: %s\n",
			id, err.Error())
	}

	// Run process until it's in a suitable state for snapshotting.
	setupSnapshotState(procPid, &ws, is64bit)

	// Save process state for future restorations.
	var procSnapshot Snapshot
	procSnapshot = makeSnapshot(procPid)

	for inputCase := range inputCases {

		writeToProc(procPid, inputCase.input)

		// Trace execution and report back interesting cases.
		curExecTrace := traceSyscalls(procPid, &ws, is64bit)
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
