package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
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
}

func isValidCSV(fileBytes []byte) bool {
	r := bytes.NewReader(fileBytes)
	reader := csv.NewReader(r)

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

func isValidJSON(fileBytes []byte) bool {
	return json.Valid(fileBytes)
}
