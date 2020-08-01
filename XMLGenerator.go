package main

import (
	"os"
	"aqwari.net/xml/xmltree"
	"fmt"
	"log"
	"io/ioutil"
	"math/rand"
	"time"
)

/*
 * elemSet is a slice of pointers to all unique elements in the tree, 
 * allowing an element to be randomly selected in O(1) time. 
 * XMLTree is the root element.  
 */
type deserializedXML struct {
	XMLTree	*xmltree.Element
	elemSet []*xmltree.Element
	description []string
}


/*
 * permXML represents an XML permutator. 
 * permXML contains two channels, toHarness and toMutator that it sends
 * TestCases to. 
 * currPerm represents the current permutation the permutator is working on
 */
type permXML struct {
	toHarness chan TestCase
	toMutator chan TestCase
	rng *rand.Rand
	currPerm *deserializedXML
}

func newXML() *deserializedXML {
	s := new(deserializedXML)
	return s
}

func newXMLPermutator(harnessChan chan TestCase, mutatorChan chan TestCase, seed int64) *permXML {
	p := new(permXML)
	p.toHarness = harnessChan
	p.toMutator = mutatorChan
	r := rand.New(rand.NewSource(seed))
	p.rng = r
	return p
}


/*
 * Reads the XML specified by file into a deserializedXML struct. 
 */
func parseXML(file string, s *deserializedXML) {
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
	elems := t.Flatten()

	// add root of the tree to elems since the Flatten method
	// doesn't do it.
	elems = append(elems, t)
	return elems
}

/*
 * randomly selects an element from the pool of elements in
 * a deserializedXML
 */
func selectElement(s *deserializedXML) *xmltree.Element {
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	i := r.Intn(len(s.elemSet))
	return s.elemSet[i]
}

/*
 * Creates a new Element which contains the same StartElement and
 * and Content as the element e. The new Element has no children. 
 */
func childlessClone(e *xmltree.Element) *xmltree.Element {
	clone := new(xmltree.Element)
	clone.StartElement = e.StartElement.Copy()
	content := make([]byte, len(e.Content))
	copy(content, e.Content)
	clone.Content = content
	clone.Children = []xmltree.Element{}
	return clone
}

/*
 * Randomly selects two elements from the element pool. 
 * One element is chosen to be the parent, the other element is
 * chosen to be the child. Multiple childless copies of the child 
 * are added to the parent. 
 */
func (s *deserializedXML) spamElementBreadthWise() {
	parent := selectElement(s)
	child := selectElement(s)

	for i := 0; i < 10; i++ {
		clone := childlessClone(child)
		parent.Children = append(parent.Children, *clone)
	}

}

/*
 * Randomly selects two elements, a parent and a child.
 * The child is recursively added to itself creating a tree.
 * This tree is then added to the parent. 
 */
func (s *deserializedXML) spamElementDepthWise() {
	parent := selectElement(s)
	child := selectElement(s)

	var root *xmltree.Element
	for i := 0; i < 10; i++ {
		root = childlessClone(child)
		root.Children = append(root.Children, *child)
		child = root
	}

	parent.Children = append(parent.Children, *root)
}

/*
 * Takes an XML file as base and permutates it. The resulting permutations
 * are converted to a TestCase and are sent across the two channels
 * to Harness and toMutator
 */
func (p *permXML) permutateInput(file string) {
	for {
		p.currPerm = newXML()
	}
}
