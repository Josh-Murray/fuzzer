package main

import (
	"io/ioutil"
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
	file string, seed int64) permutator {

	fileBytes, _ := ioutil.ReadFile(file)

	if isValidCSV(fileBytes) {
		p := newCSVPermutator(toHarness, toMutator)
		return p
	} else if isValidXML(fileBytes) {
		p := newXMLPermutator(toHarness, toMutator, seed)
		return p
	}

	if isValidJSON(file) {
		p := newJSONPermutator(toHarness, toMutator, seed)
		return p
	}

	return nil
}
