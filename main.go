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
	mutatorToHarness := make(chan TestCase)
	harnessToInteresting := make(chan TestCase)
	crashCases := make(chan TestCase)

	input, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		fmt.Println("Unable to read input file")
		return
	}

	if isValidCSV(os.Args[2]) {
		generatorToHarness := make(chan TestCase)
		go generateCSVs(generatorToHarness, os.Args[2])
		go harness(5, "./"+os.Args[1], generatorToHarness,
			harnessToInteresting, crashCases)
	}

	// create mutator threads
	for i := 0; i < 4; i++ {
		go func(i int) {
			mutator := createMutator(mutatorToHarness, int64(i))
			for {
				temp := append([]byte{}, input...)
				ts := &TestCase{temp, []string{}}
				mutator.mutate(ts)
			}
		}(i)
	}
	// create harness threads
	for i := 0; i < 5; i++ {
		go harness(i, "./"+os.Args[1], mutatorToHarness,
			harnessToInteresting, crashCases)

	}

	for crashCase := range crashCases {
		crashReport(crashCase)
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
