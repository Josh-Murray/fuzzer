package main

import (
	"log"
	"os"
	"runtime"
	"syscall"
)

/*
 * Traps on every program syscall entry/exit and records the state.
 * Function should be called after the child process has been started
 * with ptrace enabled and has had syscall.Wait4 called on it once.
 * Continues program until exit or crash: ws will be updated with the status.
 * Takes as arguments the pid of process to trace and WaitStatus to update.
 * Returns an execTrace struct identifying the execution run.
 */
func traceSyscalls(pid int, ws *syscall.WaitStatus) execTrace {
	var err error
	var regs syscall.PtraceRegs
	var curExecTrace execTrace
	for {
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
	}
}

/*
 * Harness will run the external binary specified by cmd and feed
 * it inputs from the inputCases channel. Interesting TestCases will
 * be placed in the interestCases output channel.
 */
func harness(id int, cmd string,
	inputCases <-chan TestCase,
	interestCases chan<- TestCase) {

	// List of unique execution traces for this harness.
	var uniqueTraces []execTrace

	for inputCase := range inputCases {
		var err error

		// Pipe used to pass input to binary stdin.
		procPipe, harnessPipe, err := os.Pipe()
		if err != nil {
			log.Fatalf("Harness with id %d failed to create pipe: %s\n",
				id, err.Error())
		}

		var pAttr syscall.ProcAttr
		pAttr.Sys = &syscall.SysProcAttr{Ptrace: true}

		// Ignore process stdout/stderr.
		pAttr.Files = make([]uintptr, 1)
		pAttr.Files[0] = procPipe.Fd()

		// Lock OS thread as per syscall.SysProcAttr documentation.
		runtime.LockOSThread()

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

		_, err = harnessPipe.Write(inputCase.input)
		if err != nil {
			log.Fatalf("Harness with id %d failed to write to program: %s\n",
				id, err.Error())
		}

		// Process may need pipe closed to continue.
		err = harnessPipe.Close()
		if err != nil {
			log.Printf("Harness with id %d failed to manually close stdin pipe.\n",
				id)
		}

		// Trace execution and report back interesting cases.
		curExecTrace := traceSyscalls(procPid, &ws)
		runtime.UnlockOSThread()
		if isUniqueTrace(curExecTrace, uniqueTraces) {
			uniqueTraces = append(uniqueTraces, curExecTrace)
			// This channel is currently not used, leading to deadlock
			// if given input here.
			//interestCases <- inputCase
		}

		// Perform process cleanup on abort.
		var aborted bool = false
		if ws.StopSignal() == syscall.SIGABRT {
			aborted = true

			err = syscall.PtraceDetach(procPid)
			if err != nil {
				log.Printf("Harness with id %d failed to detach: %s\n",
					id, err.Error())
			}

			_, err = syscall.Wait4(procPid, &ws, syscall.WALL, nil)
			if err != nil {
				log.Fatalf("Harness with id %d failed to wait for tracee "+
					"on abort: %s\n", id, err.Error())
			}

		}

		// Report segfaults and ignore other exit causes.
		if ws.StopSignal() == syscall.SIGSEGV && aborted == false {
			log.Printf("Harness with id %d crashed process with pid %d\n",
				id, procPid)
			crashReport(inputCase)
		}

		procPipe.Close()
	}
}

/*
 * Creates a "bad.txt" file in the current directory containing
 * the input inside crashCase
 */
func crashReport(crashCase TestCase) {
	var doExit bool = true

	f, err := os.OpenFile("bad.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	// Log the crashing input on any file operation failure.
	if err != nil {
		log.Println("Failed to create crash output file. Crashing output:")
		log.Println(string(crashCase.input))
		return
	}

	nWritten, err := f.Write(crashCase.input)
	// Log crash output on failed or incomplete writes.
	if err != nil || nWritten != len(crashCase.input) {
		log.Println("Failed to write output to crash file. Crashing output:")
		log.Println(string(crashCase.input))
		// Continue execution to try hit another crash.
		doExit = false
	}

	err = f.Close()
	if err != nil {
		log.Fatal("crashReport failed to close the file")
	}

	// Stop execution on first bad output hit unless there was an error
	// in file generation.
	if doExit {
		os.Exit(0)
	}
}
