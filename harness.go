package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

/*
 * Harness will run the external binary specified by cmd and feed
 * it inputs from the inputCases channel. Interesting TestCases will
 * be placed in the interestCases output channel.
 */
func harness(id int, cmd string,
	inputCases <-chan TestCase,
	interestCases chan<- TestCase) {

	for inputCase := range inputCases {
		var err error
		procCmd := exec.Command(cmd)
		procStdin, err := procCmd.StdinPipe()
		if err != nil {
			log.Fatalf("Harness with id %d failed to connect stdin pipe: %s\n",
				id, err.Error())
		}

		procStderr, err := procCmd.StderrPipe()
		if err != nil {
			log.Fatalf("Harness with id %d failed to connect stderr pipe: %s\n",
				id, err.Error())

		}

		err = procCmd.Start()
		if err != nil {
			log.Fatalf("Harness with id %d failed to start program: %s\n",
				id, err.Error())
		}

		procPid := procCmd.Process.Pid
		_, err = procStdin.Write(inputCase.input)
		if err != nil {
			log.Fatalf("Harness with id %d failed to write to program: %s\n",
				id, err.Error())
		}

		// Process may need pipe closed to continue.
		err = procStdin.Close()
		if err != nil {
			log.Printf("Harness with id %d failed to manually close stdin pipe.\n",
				id)
		}

		procErr, err := ioutil.ReadAll(procStderr)
		if err != nil {
			log.Fatalf("Harness with id %d failed to read program stderr\n",
				id)
		}

		procCmd.Wait()

		// Check program stderr for ASAN crash output
		var asanCrash bool = false
		if strings.Contains(string(procErr), "ERROR: AddressSanitizer") {
			asanCrash = true
		}

		// Report segfaults and ASAN crashes, ignore other exit causes.
		waitStatus := procCmd.ProcessState.Sys().(syscall.WaitStatus)
		if waitStatus.Signal() == syscall.SIGSEGV || asanCrash == true {
			log.Printf("Harness with id %d crashed process with pid %d\n",
				id, procPid)
			crashReport(inputCase)
		}
	}
}

/*
 * Creates a "bad.txt" file in the current directory containing
 * the input inside crashCase
 */
func crashReport(crashCase TestCase) {
	f, err := os.OpenFile("bad.txt", os.O_WRONLY|os.O_CREATE, 0644)
	// Log the crashing input on any file operation failure.
	if err != nil {
		log.Println("Failed to create crash output file. Crashing output:")
		log.Println(string(crashCase.input))
	}

	nWritten, err := f.Write(crashCase.input)
	if err != nil {
		log.Println("Failed to write output to crash file. Crashing output:")
		log.Println(string(crashCase.input))
	}

	// Also log crash output on incomplete writes.
	if nWritten != len(crashCase.input) {
		log.Println("Failed to write full output to crash file. Crashing output:")
		log.Println(string(crashCase.input))
	}

	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}
}
