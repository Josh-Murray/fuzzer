package main

import (
	"fmt"
)

type TestCase struct {
	input   []byte
	changes []string
}

func (ts TestCase) printTestCase() {
	fmt.Println(string(ts.input))
	for index, value := range ts.changes {
		fmt.Printf("Change %d: %s", index, value)
	}
}
func (ts TestCase) printInput() {
	fmt.Println(string(ts.input))
}
func (ts TestCase) printChanges() {
	for index, value := range ts.changes {
		fmt.Printf("Change %d: %s", index, value)
	}
}
