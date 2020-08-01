package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
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

	input, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		fmt.Println("Unable to read input file")
		return
	}

	go tsPrinter(mutatorToHarness)

	mutator := createMutator(mutatorToHarness, int64(1))
	temp := append([]byte{}, input...)
	ts := &TestCase{temp, []string{}}

	for i := 0; i < 4; i++ {
		mutator.testMutate(ts)
	}
	time.Sleep(100 * time.Millisecond)
}

func tsPrinter(chanIn chan TestCase) {
	for i := 0; i < 4; i++ {
		currentCase := <-chanIn
		fmt.Printf("---------------- %2d ----------------\n", i)
		currentCase.printTestCase()
		fmt.Printf("---------------- --- ----------------\n")
	}
}
