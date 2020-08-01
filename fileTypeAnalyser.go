package main

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func isValidXML(data []byte) bool {
	err := xml.Unmarshal(data, new(interface{}))
	if err == nil {
		return true
	}
	return false
	//return xml.Unmarshal(data, new(interface{})) != nil
}

func isValidCSV(file string) bool {
	csvFile, _ := os.Open(file)
	reader := csv.NewReader(csvFile)

	for {
		_, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(err)
			return false
		}
	}
	return true
}

func isValidJSON(file string) bool {
	data, err := ioutil.ReadFile(file)
	check(err)
	return json.Valid(data)
}

func detectType(file string) {

	f, err := os.Open(file)
	check(err)
	f.Close()

	data, err := ioutil.ReadFile(file)
	check(err)
	//fmt.Println(data)
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
		fmt.Println(xml.Unmarshal(data, new(interface{})))
	}

	if isValidCSV(file) {
		fmt.Print(" valid csv")
	} else {
		fmt.Print("invalid csv")
	}
}

//func main() {
//	detectType("mutator.go")
//}
