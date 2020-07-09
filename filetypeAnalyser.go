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
	err := xml.Unmarshal(data, new(interface{}))
	if err == nil {
		return true
	}
	return false
	//return xml.Unmarshal(data, new(interface{})) != nil
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

}

func main() {
	detectType("cd_catalog.xml")
}
