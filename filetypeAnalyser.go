package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}

func isValidXML(data []byte) bool {
	return xml.Unmarshal(data, new(interface{})) != nil
}

func detectType(file string) {

	f, err := os.Open(file)
	check(err)
	f.Close()

	data, err := ioutil.ReadFile(file)
	check(err)

	ext := filepath.Ext(file)
	fmt.Print("Extension is: ", ext)

	if json.Valid(data) {
		fmt.Print(" valid json")
	} else {
		fmt.Print(" invalid json")
	}

	if isValidXML(data) {
		fmt.Print(" valid xml")
	} else {
		fmt.Print(" invalid xml")
	}

}
