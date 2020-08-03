package main

import (
	"aqwari.net/xml/xmltree"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
)

/*
 * elemSet is a slice of pointers to all unique elements in the tree,
 * allowing an element to be randomly selected in O(1) time.
 * XMLTree is the root element.
 */
type deserializedXML struct {
	XMLTree     *xmltree.Element
	elemSet     []*xmltree.Element
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
	rng       *rand.Rand
	currPerm  *deserializedXML
}

/*
 * creates a new deserializedXML struct.
 * Initialises the XMLTree field with parseXML.
 * Initializes the elemSet field with getElemSet.
 */
func newXML(file string) *deserializedXML {
	s := new(deserializedXML)
	parseXML(file, s)
	s.elemSet = getElemSet(s.XMLTree)
	return s
}

/*
 * converts the permutation into a byte slice representing the
 * XML content
 */
func (perm *deserializedXML) marshal() []byte {
	return xmltree.Marshal(perm.XMLTree)
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
 * randomly selects an element from the set of elements elemSet in
 * the current permutation, currPerm
 */
func (p *permXML) selectElement() *xmltree.Element {
	i := p.rng.Intn(len(p.currPerm.elemSet))
	return p.currPerm.elemSet[i]
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
 * Two elements are randomly selected. One acts as the parent, the
 * Other acts as the child.
 * Multiple childless copies of the child
 * are added to the parent.
 */
func spamElementBreadthWise(p *permXML) {

	parent := p.selectElement()
	child := p.selectElement()

	for i := 0; i < 25000; i++ {
		clone := childlessClone(child)
		parent.Children = append(parent.Children, *clone)
	}

	p.currPerm.addToDesc("Spamming Element Breadth Wise")

}

/*
 * Two elements are randomly selected. One acts as the parent, the
 * other acts as the child
 * The child is recursively added to itself creating a tree.
 * This tree is then added to the parent.
 */
func spamElementDepthWise(p *permXML) {

	parent := p.selectElement()
	child := p.selectElement()

	var root *xmltree.Element

	for i := 0; i < 25000; i++ {
		root = childlessClone(child)
		root.Children = append(root.Children, *child)
		child = root
	}

	parent.Children = append(parent.Children, *root)

	p.currPerm.addToDesc("Spamming Element Depth Wise")
}

func (p *permXML) generateTestCase() {
	ts := TestCase{}
	ts.changes = append(ts.changes, p.currPerm.description...)
	content := p.currPerm.marshal()
	ts.input = content
	p.toHarness <- ts
	p.toMutator <- ts
}

func (perm *deserializedXML) addToDesc(change string) {
	perm.description = append(perm.description, change)
}

/* Doesn't perform any permutation on the deserializedXML */
func plainXML(perm *deserializedXML) {
	perm.addToDesc("Original input")
}

/*
 * Takes an XML file as base and permutates it. The resulting permutations
 * are converted to a TestCase and are sent across the two channels
 * to Harness and toMutator
 */
func (p *permXML) permutateInput(file string) {
	for {
		p.currPerm = newXML(file)
		plainXML(p.currPerm)
		p.generateTestCase()

		p.currPerm = newXML(file)
		spamElementBreadthWise(p)
		p.generateTestCase()

		p.currPerm = newXML(file)
		spamElementDepthWise(p)
		p.generateTestCase()
	}
}
