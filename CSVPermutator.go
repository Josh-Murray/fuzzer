package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

/*
 * parsedCSV contains parsed data from a CSV file
 */
type parsedCSV struct {
	lines         [][]string
	rows, columns int
	description   []string
}

/* permCSV represents a CSV permutator. 
 * permCSV conforms to the permutator interface in that it implements
 * the method permutateInput. 
 * permCSV contains two channels, toHarness and toMutator, that it sends
 * TestCases to.
 * currPerm represents the current permutation the permutator is working on.
 */
type permCSV struct {
	toHarness     chan TestCase
	toMutator	chan TestCase
	currPerm *parsedCSV
}


/*
 * Create a new parsedCSV struct
 */
func newCSV() *parsedCSV {
	c := new(parsedCSV)
	return c
}

/*
 * TODO: add Description
 */
func newCSVPermutator(harnessChan chan TestCase, mutatorChan chan TestCase) *permCSV{
	p := new(permCSV)
	p.toHarness = harnessChan
	p.toMutator = mutatorChan
	return p
}

/*
 * parse the CSV specified by file into a parsedCSV struct
 */
func parse(file string, c *parsedCSV) {

	csvFile, _ := os.Open(file)
	reader := csv.NewReader(csvFile)
	defer csvFile.Close()

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		c.lines = append(c.lines, line)

		c.columns = len(line)
	}

	c.rows = len(c.lines)
	operation := fmt.Sprintf("Read in CSV from: %s with %d rows, %d cols \n", file, c.rows, c.columns)
	c.addToDesc(operation)
}

/*
 * Add a change to the description in a parsedCSV
 */
func (s *parsedCSV) addToDesc(change string) {
	s.description = append(s.description, change)
}

/*
 * Display the current CSV in STDOUT
 */
func (s *parsedCSV) display() {
	fmt.Printf("Description: %s\n", s.description)
	fmt.Printf("Rows: %d, Columns: %d\n", s.rows, s.columns)
	fmt.Print("Column \t\t|")
	for i := 0; i < s.columns; i++ {
		fmt.Printf("\t%d\t", i)
	}
	fmt.Println("\n\t\t", strings.Repeat("-", s.columns*15))
	for i, line := range s.lines {
		fmt.Printf("row %2d: \t|", i)
		for _, c := range line {
			//fmt.Printf(" %d> %4s   <", j, c)
			fmt.Printf("\t%2s\t", c)

		}
		fmt.Println("")
	}
}

/*
 * Delete row u
 */
func (s *parsedCSV) deleteRow(u int) {
	if u < s.rows {
		s.lines = append(s.lines[:u], s.lines[u+1:]...)
		s.addToDesc(fmt.Sprintf("Removed row %d from CSV\n", u))
		s.rows--
	} else {
		s.addToDesc(fmt.Sprintf("No row %d to remove from CSV\n", u))
	}

}

/*
 * Delete column u
 */
func (s *parsedCSV) deleteCol(u int) {

	if u < s.columns {
		for i, line := range s.lines {
			s.lines[i] = append(line[:u], line[u+1:]...)
		}
		s.addToDesc(fmt.Sprintf("Removed column %d from CSV\n", u))
		s.columns--
	} else {
		s.addToDesc(fmt.Sprintf("No Column %d to delete from CSV\n", u))
	}

}

/*
 * Insert a row rowToAdd into location l in the parsedCSV c
 */
func addRow(l int, rowToAdd []string, c *parsedCSV) {
	var operation string
	if l < c.rows {
		if len(rowToAdd) == c.columns {
			end := c.lines[l:]
			beginning := append(c.lines[:l], rowToAdd)
			c.lines = append(beginning, end...)
			operation = fmt.Sprintln("added csv row ", l, " content ", rowToAdd)

			c.rows++
		} else {
			operation += "Number of columns in the row being added does not match"
		}
	} else {
		operation += fmt.Sprintf("%d is not a valid location to insert a row\n", l)
	}

	c.addToDesc(operation)
}

/*
 * Insert a column colToAdd into location l
 */
func addColumn(l int, colToAdd []string, currPerm *parsedCSV) {
	var operation string
	if l > currPerm.columns {
		l = currPerm.columns
	}
	if len(colToAdd) == currPerm.rows {
		for i := 0; i < currPerm.rows; i++ {
			end := currPerm.lines[i][l:]
			beginning := append(currPerm.lines[i][:l], colToAdd[i])
			currPerm.lines[i] = append(beginning, end...)

		}
		operation = fmt.Sprintln("added csv col ", l, " content ", colToAdd)

		currPerm.columns++
	}
	currPerm.addToDesc(operation)

}

/*
 * Return a []string representing column c in currPerm
 */
func getCol(c int, currPerm *parsedCSV) []string {
	col := []string{}
	for _, row := range currPerm.lines {
		col = append(col, row[c])
	}
	return col
}

/*
 * Return a deep copy of row r
 * TODO: Investigate why this isnt working
 */
func getrRow(r int, currPerm *parsedCSV) []string {
	cpy := []string{}
	for _, str := range currPerm.lines[r] {
		cpy = append(cpy, fmt.Sprintf("%s", str))
	}
	return cpy
}

/*
 * Insert a duplicate of column c 
 */
func copyCol(c int, currPerm *parsedCSV) {
	if c < currPerm.columns {
		col := getCol(c, currPerm)
		addColumn(c, col, currPerm)

		currPerm.addToDesc(fmt.Sprintln("coppied csv column", c))
	}
}

/*
 * Insert a duplicate of row r
 */
func copyRow(r int, c *parsedCSV) {
	if r < c.rows {
		row := getrRow(r, c)
		addRow(r, row, c)
		c.addToDesc(fmt.Sprintln("coppied csv row", r))
	}
}

/*
 * convert s into a string representing the content of the CSV
 */
func (s *parsedCSV) flatten() string {
	o := ""
	for _, row := range s.lines {
		for j, col := range row {
			o += fmt.Sprint(col)
			if j < s.columns-1 {
				o += fmt.Sprint(",")
			}
		}
		o += fmt.Sprint("\n")
	}
	return fmt.Sprint(o)
}

/*
 * parse a string into the parsedCSV struct
 */
func (s *parsedCSV) expand(in string) {
	// TODO: can probably refactor this to reduce code duplication with readCSV
	reader := csv.NewReader(strings.NewReader(in))
	s.lines = [][]string{}
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		s.lines = append(s.lines, line)

		s.columns = len(line)
	}
	s.rows = len(s.lines)
}

/*
 * Flatten then reexpand inplace
 */
func (s *parsedCSV) reExpand() {
	s.expand(s.flatten())
}

/*
 * Insert a blank row into position r in the parsedCSV c
 */
func addBlankRow(r int, currPerm *parsedCSV) {
	blankRow := make([]string, currPerm.columns)
	addRow(r, blankRow, currPerm)
}

/*
 * Insert a blank column into position c
 */
func addBlankCol(c int, currPerm *parsedCSV) {
	blankCol := make([]string, currPerm.rows)
	addColumn(c, blankCol, currPerm)
}

/*
 * Deep copy (hopefully) the current permutation currPerm in
 * the permutator p into a TestCase.
 * Sends the resulting TestCase across the two channels
 * in the permutator p, toHarness and toMutator.
 */
func generateTestCase(p *permCSV) {
	ts := TestCase{}
	ts.changes = append(ts.changes, p.currPerm.description...)
	content := p.currPerm.flatten()
	ts.input = append(ts.input, content...)

	p.toHarness <- ts
	p.toMutator <- ts
}


/*
 * Teses making lots of copies of a row. If copies is true then copy the last row
 * otherwise add blank rows
 */
func spamRows(copies bool, currPerm *parsedCSV) {
	if copies {
		currPerm.addToDesc("Spamming blank CSV rows")
	} else {
		currPerm.addToDesc("Spamming copies CSV rows")
	}

	//TODO: abstract magic numbers
	for i := 1; i < 4096; i++ {
		lastRow := currPerm.rows - 1
		if copies {
			copyRow(lastRow, currPerm)
		} else {
			addBlankRow(lastRow, currPerm)
		}
	}

}

/*
 * Teses making lots of copies of a column. If copies is true then copy the last row
 * otherwise add blank rows
 */
func spamCols(copies bool, currPerm *parsedCSV) {
	if copies {
		currPerm.addToDesc("Spamming blank CSV cols")
	} else {
		currPerm.addToDesc("Spamming copies CSV cols")
	}

	//TODO: abstract magic numbers
	for i := 1; i < 4096; i++ {
		lastColumn := currPerm.columns - 1
		if copies {
			copyCol(lastColumn, currPerm)
		} else {
			addBlankCol(lastColumn, currPerm)
		}

	}

}

/*
 * Blank out the csv
 */
func blankCSV(currPerm *parsedCSV) {

	blankRow := make([]string, currPerm.columns)

	newPerm := newCSV()
	newPerm.addToDesc("Blank CSV with original number of rows and columns")

	for i := 0; i < currPerm.rows; i++ {
		addRow(1, blankRow, newPerm)
	}

	/* set the current permutation as the new permutation
	 * with the blanked out rows 
	 */
	currPerm = newPerm

}

/*
 * Doesn't perform any permutation on parsedCSV. 
 * Simply converts parsedCSV to a TestCase and sends it to the
 * two channels toHarness and toMutator. 
 */
func plainCSV(currPerm *parsedCSV) {
	currPerm.addToDesc("Initial Input")
}

/*
 * Take a CSV file as base and permute variations, Converting
 * the permutations to a TestCase and sending it across the two
 * channels toHarness and toMutator
 */
func (p *permCSV) permutateInput(file string) {

	p.currPerm = newCSV()
	parse(file, p.currPerm)
	plainCSV(p.currPerm)
	generateTestCase(p)

	p.currPerm = newCSV()
	parse(file, p.currPerm)
	blankCSV(p.currPerm)
	generateTestCase(p)

	p.currPerm = newCSV()
	parse(file, p.currPerm)
	spamRows(false, p.currPerm)
	generateTestCase(p)

	p.currPerm = newCSV()
	parse(file, p.currPerm)
	spamRows(true, p.currPerm)
	generateTestCase(p)

	p.currPerm = newCSV()
	parse(file, p.currPerm)
	spamCols(false, p.currPerm)
	generateTestCase(p)

	p.currPerm = newCSV()
	parse(file, p.currPerm)
	spamCols(true, p.currPerm)
	generateTestCase(p)

}


