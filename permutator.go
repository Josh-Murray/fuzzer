package main


import (

)

type permutator interface {
	permutateInput(file string)
}

/* 
 * Determines the if input contained in the file is in 
 * CSV, XML, JSON or plaintext format in the default case. 
 * returns a struct that conforms to the permuator interface. 
 * TODO: Add XML and JSON to createPermutator. 
 */
func createPermutator(toHarness chan TestCase, toMutator chan TestCase,
		file string) permutator {

	if isValidCSV(file) {
		p := newCSVPermutator(toHarness, toMutator)
		return p
	}

	return nil
}

