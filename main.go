package main

import (
	"fmt"
	"io/ioutil"
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

	input, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		fmt.Println("Unable to read input file")
		return
	}

	if isValidCSV(os.Args[2]) {
		generatorToHarness := make(chan TestCase)
		go generateCSVs(generatorToHarness, os.Args[2])
		go harness(5, "./"+os.Args[1], generatorToHarness, harnessToInteresting)
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
	for i := 0; i < 4; i++ {
		go harness(i, "./"+os.Args[1], mutatorToHarness, harnessToInteresting)

	}

	harness(4, "./"+os.Args[1], mutatorToHarness, harnessToInteresting)
}
