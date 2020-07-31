package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

/*
 * deserializedCSV contains parsed data from a CSV file.
 * permutation functions such as spamRows and spamCols
 * operate on the deserializedCSV struct and manipulate the
 * data within it. The description field in the deserializedCSV
 * details the operations that the permutation functions
 * have performed on it.
 */
type deserializedCSV struct {
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
	toHarness chan TestCase
	toMutator chan TestCase
	currPerm  *deserializedCSV
}

/*
 * Create a new deserializedCSV struct
 */
func newCSV() *deserializedCSV {
	c := new(deserializedCSV)
	return c
}

/*
 * TODO: add Description
 */
func newCSVPermutator(harnessChan chan TestCase, mutatorChan chan TestCase) *permCSV {
	p := new(permCSV)
	p.toHarness = harnessChan
	p.toMutator = mutatorChan
	return p
}

/*
 * parse the CSV specified by file into a deserializedCSV struct.
 */
func parse(file string, perm *deserializedCSV) {

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
		perm.lines = append(perm.lines, line)

		perm.columns = len(line)
	}

	perm.rows = len(perm.lines)
	operation := fmt.Sprintf("Read in CSV from: %s with %d rows, %d cols \n", file, perm.rows, perm.columns)
	perm.addToDesc(operation)
}

/*
 * Add a change to the description in a deserializedCSV
 */
func (perm *deserializedCSV) addToDesc(change string) {
	perm.description = append(perm.description, change)
}

/*
 * Display the current CSV in STDOUT
 */
func (s *deserializedCSV) display() {
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
 * Deletes the row r from the permutation perm
 */
func deleteRow(r int, perm *deserializedCSV) {
	if r < perm.rows {
		perm.lines = append(perm.lines[:r], perm.lines[r+1:]...)
		perm.addToDesc(fmt.Sprintf("Removed row %d from CSV\n", r))
		perm.rows--
	} else {
		perm.addToDesc(fmt.Sprintf("No row %d to remove from CSV\n", r))
	}

}

/*
 * Deletes col c from the permutation perm
 */
func deleteCol(c int, perm *deserializedCSV) {

	if c < perm.columns {
		for i, line := range perm.lines {
			perm.lines[i] = append(line[:c], line[c+1:]...)
		}
		perm.addToDesc(fmt.Sprintf("Removed column %d from CSV\n", c))
		perm.columns--
	} else {
		perm.addToDesc(fmt.Sprintf("No Column %d to delete from CSV\n", c))
	}

}

/*
 * Insert a row rowToAdd into location l in the current permutation
 * perm
 */
func addRow(l int, rowToAdd []string, perm *deserializedCSV) {
	var operation string
	if l < perm.rows {
		if len(rowToAdd) == perm.columns {
			end := perm.lines[l:]
			beginning := append(perm.lines[:l], rowToAdd)
			perm.lines = append(beginning, end...)
			operation = fmt.Sprintln("added csv row ", l, " content ", rowToAdd)

			perm.rows++
		} else {
			operation += "Number of columns in the row being added does not match"
		}
	} else {
		operation += fmt.Sprintf("%d is not a valid location to insert a row\n", l)
	}

	perm.addToDesc(operation)
}

/*
 * Insert a column colToAdd into location l
 */
func addColumn(l int, colToAdd []string, perm *deserializedCSV) {
	var operation string
	if l > perm.columns {
		l = perm.columns
	}
	if len(colToAdd) == perm.rows {
		for i := 0; i < perm.rows; i++ {
			end := perm.lines[i][l:]
			beginning := append(perm.lines[i][:l], colToAdd[i])
			perm.lines[i] = append(beginning, end...)

		}
		operation = fmt.Sprintln("added csv col ", l, " content ", colToAdd)

		perm.columns++
	}
	perm.addToDesc(operation)

}

/*
 * Return a []string representing column c in perm
 */
func getCol(c int, perm *deserializedCSV) []string {
	col := []string{}
	for _, row := range perm.lines {
		col = append(col, row[c])
	}
	return col
}

/*
 * Return a deep copy of row r
 * TODO: Investigate why this isnt working
 */
func getrRow(r int, perm *deserializedCSV) []string {
	cpy := []string{}
	for _, str := range perm.lines[r] {
		cpy = append(cpy, fmt.Sprintf("%s", str))
	}
	return cpy
}

/*
 * Insert a duplicate of column c
 */
func copyCol(c int, perm *deserializedCSV) {
	if c < perm.columns {
		col := getCol(c, perm)
		addColumn(c, col, perm)

		perm.addToDesc(fmt.Sprintln("coppied csv column", c))
	}
}

/*
 * Insert a duplicate of row r
 */
func copyRow(r int, perm *deserializedCSV) {
	if r < perm.rows {
		row := getrRow(r, perm)
		addRow(r, row, perm)
		perm.addToDesc(fmt.Sprintln("coppied csv row", r))
	}
}

/*
 * convert the current permutation perm
 * into a string representing the content of the CSV
 */
func (perm *deserializedCSV) flatten() string {
	res := ""
	for _, row := range perm.lines {
		for j, col := range row {
			res += fmt.Sprint(col)
			if j < perm.columns-1 {
				res += fmt.Sprint(",")
			}
		}
		res += fmt.Sprint("\n")
	}
	return fmt.Sprint(res)
}

/*
 * Insert a blank row into position r in the deserializedCSV c
 */
func addBlankRow(r int, perm *deserializedCSV) {
	blankRow := make([]string, perm.columns)
	addRow(r, blankRow, perm)
}

/*
 * Insert a blank column into position c
 */
func addBlankCol(c int, perm *deserializedCSV) {
	blankCol := make([]string, perm.rows)
	addColumn(c, blankCol, perm)
}

/*
 * Deep copy (hopefully) the current permutation perm in
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
func spamRows(copies bool, perm *deserializedCSV) {
	if copies {
		perm.addToDesc("Spamming blank CSV rows")
	} else {
		perm.addToDesc("Spamming copies CSV rows")
	}

	//TODO: abstract magic numbers
	for i := 1; i < 4096; i++ {
		lastRow := perm.rows - 1
		if copies {
			copyRow(lastRow, perm)
		} else {
			addBlankRow(lastRow, perm)
		}
	}

}

/*
 * Teses making lots of copies of a column. If copies is true then copy the last row
 * otherwise add blank rows
 */
func spamCols(copies bool, perm *deserializedCSV) {
	if copies {
		perm.addToDesc("Spamming blank CSV cols")
	} else {
		perm.addToDesc("Spamming copies CSV cols")
	}

	//TODO: abstract magic numbers
	for i := 1; i < 4096; i++ {
		lastColumn := perm.columns - 1
		if copies {
			copyCol(lastColumn, perm)
		} else {
			addBlankCol(lastColumn, perm)
		}

	}

}

/*
 * Blank out the csv
 */
func blankCSV(perm *deserializedCSV) {

	blankRow := make([]string, perm.columns)

	newPerm := newCSV()
	newPerm.addToDesc("Blank CSV with original number of rows and columns")

	for i := 0; i < perm.rows; i++ {
		addRow(1, blankRow, newPerm)
	}

	/* set the current permutation as the new permutation
	 * with the blanked out rows
	 */
	perm = newPerm

}

/*
 * Doesn't perform any permutation on deserializedCSV.
 * Simply converts deserializedCSV to a TestCase and sends it to the
 * two channels toHarness and toMutator.
 */
func plainCSV(perm *deserializedCSV) {
	perm.addToDesc("Initial Input")
}

/*
 * Take a CSV file as base and permute variations, Converting
 * the permutations to a TestCase and sending it across the two
 * channels toHarness and toMutator
 */
func (p *permCSV) permutateInput(file string) {

	for {

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

}
