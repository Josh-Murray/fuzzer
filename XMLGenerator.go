package main

import (
	"os"
	"aqwari.net/xml/xmltree"

)

type mXMLHolder struct {
	XMLTree	*xmltree.Element
	description []string
}

func createXMLHolder(initialDescription string) mXMLHolder {
	s := mXMLHolder{}
	s.description = append(s.description, initialDescription)
	return s
}


/*
 * Reads the XML specified by file into the XMLHolder. 
 */
func (s *mXMLHolder) read(file string) {
	xmlFile, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer xmlFile.Close()

	xmlBytes, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		log.Fatal(err)
	}

	s.XMLTree, err := xmlTree.Parse(xmlBytes)
	if err != nil {
		log.Fatal(err)
	}

	operation := fmt.Sprintf("Read in XML from: %s\n", file)
	s.description = append(s.description, operation)

}


