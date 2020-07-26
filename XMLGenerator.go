package main

import (
	"os"
	"aqwari.net/xml/xmltree"
	"fmt"
	"log"
	"io/ioutil"
	"math/rand"
)


/*
 * elemSet is a slice of pointers to all elements in the tree, 
 * allowing an element to be randomly selected in O(1) time. 
 * XMLTree is the root element.  
 */
type mXMLHolder struct {
	XMLTree	*xmltree.Element
	elemSet []*xmltree.Element
	description []string
}

func createXMLHolder(description string) mXMLHolder {
	s := mXMLHolder{}
	s.description = append(s.description, description)
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

	s.XMLTree, err = xmltree.Parse(xmlBytes)
	if err != nil {
		log.Fatal(err)
	}

	operation := fmt.Sprintf("Read in XML from: %s\n", file)
	s.description = append(s.description, operation)

}

/*
 * Returns a slice of all Elements in the tree t
 */
func getElemSet(t *xmltree.Element) []*xmltree.Element {
	queue := []*xmltree.Element{}
	queue = append(queue, t)
	res := []*xmltree.Element{}
	return BFSUtil(queue, res)
}

/*
 * a utility function to perform BFS given a queue of Elements to visit. 
 * BFSUtil returns a slice of all Elements visited. 
 */
func BFSUtil(queue []*xmltree.Element, res []*xmltree.Element) []*xmltree.Element {
	// all elements have been visited if the queue is empty. 	
	if len(queue) == 0 {
		return res
	}

	// visit element at the front of the queue and add it to res
	// to mark it as visited.
	elem := queue[0]
	res = append(res, elem)

	// add all elem's children to the queue to mark them as
	// still to be visited. 
	for i := range elem.Children {
		queue = append(queue, &elem.Children[i])
	}

	// visit next element in the queue. 
	return BFSUtil(queue[1:],res)
}




