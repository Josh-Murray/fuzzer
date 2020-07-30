package main


import (

)

type permutator interface {
	permutateInput(h chan<- TestCase, m chan<- TestCase, file string)
}

/* 
 * Determines the if input contained in the file is in 
 * CSV, XML, JSON or plaintext format in the default case. 
 * returns a struct that conforms to the permuator interface. 
 * TODO: Add XML and JSON to createPermutator. 
 */
func createPermutator(file string) permutator {
	if isValidCSV(file) {
		p := newCSV("inital input")
		return p
	}

	return nil
}

