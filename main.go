package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

/*
 * Accepts an executable file and an input file to run it with.
 */

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage:", os.Args[0], "<binary>", "<input file>")
		return
	}

	// create channels for mutator and harness
	inputToHarness := make(chan TestCase, 1000)
	inputToMutator := make(chan TestCase, 1000)
	harnessToInteresting := make(chan TestCase, 1000)
	crashCases := make(chan TestCase)

	input, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		log.Fatal("Unable to read input file")
	}

	// create permutator threads
	startPermutators(inputToHarness, inputToMutator, os.Args[2])

	// create mutator threads
	startMutators(inputToHarness, inputToMutator, input)

	// create feedback thread
	startFeedback(harnessToInteresting, inputToHarness)
	// create harness threads
	startHarnesses(os.Args[1], inputToHarness, harnessToInteresting, crashCases)

	for crashCase := range crashCases {
		crashReport(crashCase)
	}
}

func startMutators(outChan chan TestCase, inChan chan TestCase, input []byte) {
	for i := 0; i < 3; i++ {
		go func(i int) {
			mutator := createMutator(outChan, inChan, input, int64(i))
			for {
				mutator.mutate()
			}
		}(i)
	}

}

func startPermutators(toHarness chan TestCase, toMutator chan TestCase, file string) {
	for i := 0; i < 4; i++ {
		go func(i int) {
			permutator := createPermutator(toHarness, toMutator, file, int64(i))
			permutator.permutateInput(file)
		}(i)
	}
}

func startHarnesses(binary string, inChan chan TestCase,
	outChan chan TestCase, crashChan chan TestCase) {

	for i := 0; i < 4; i++ {
		go harness(i, "./"+binary, inChan, outChan, crashChan)

	}

}

func startFeedback(harnessToInteresting chan TestCase, inputToHarness chan TestCase){
	mutator := createMutator(inputToHarness, harnessToInteresting, []byte{}, int64(0))
	go mutator.feedbackLoop()
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
