package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

type mCSVHolder struct {
	lines         [][]string
	rows, columns int
	description   []string
}

/*
 * Create a new CSVHolder initialised with the initial description
 *
 */
func newCSV(initialDescription string) mCSVHolder {
	s := mCSVHolder{}
	s.description = append(s.description, initialDescription)
	return s
}

/*
 * read the CSV specified by file into the CSVHolder
 */
func (s *mCSVHolder) read(file string) {
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
		s.lines = append(s.lines, line)

		s.columns = len(line)
	}
	s.rows = len(s.lines)
	operation := fmt.Sprintf("Read in CSV from: %s with %d rows, %d cols \n", file, s.rows, s.columns)
	s.description = append(s.description, operation)
}

/*
 * Display the current CSV in STDOUT
 */
func (s *mCSVHolder) display() {
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
func (s *mCSVHolder) deleteRow(u int) {
	if u < s.rows {
		s.lines = append(s.lines[:u], s.lines[u+1:]...)
		s.description = append(s.description, fmt.Sprintf("Removed row %d from CSV\n", u))
		s.rows--
	} else {
		s.description = append(s.description, fmt.Sprintf("No row %d to remove from CSV\n", u))
	}

}

/*
 * Delete column u
 */
func (s *mCSVHolder) deleteCol(u int) {

	if u < s.columns {
		for i, line := range s.lines {
			s.lines[i] = append(line[:u], line[u+1:]...)
		}
		s.description = append(s.description, fmt.Sprintf("Removed column %d from CSV\n", u))
		s.columns--
	} else {
		s.description = append(s.description, fmt.Sprintf("No Column %d to delete from CSV\n", u))
	}

}

/*
 * Insert a row rowToAdd into location l
 */
func (s *mCSVHolder) addRow(l int, rowToAdd []string) {
	var operation string
	if l < s.rows {
		if len(rowToAdd) == s.columns {
			end := s.lines[l:]
			beginning := append(s.lines[:l], rowToAdd)
			s.lines = append(beginning, end...)
			operation = fmt.Sprintln("added csv row ", l, " content ", rowToAdd)

			s.rows++
		} else {
			operation += "Number of columns in the row being added does not match"
		}
	} else {
		operation += fmt.Sprintf("%d is not a valid location to insert a row\n", l)
	}

	s.description = append(s.description, operation)
}

/*
 * Insert a column colToAdd into location l
 */
func (s *mCSVHolder) addColumn(l int, colToAdd []string) {
	var operation string
	if l > s.columns {
		l = s.columns
	}
	if len(colToAdd) == s.rows {
		for i := 0; i < s.rows; i++ {
			end := s.lines[i][l:]
			beginning := append(s.lines[i][:l], colToAdd[i])
			s.lines[i] = append(beginning, end...)

		}
		operation = fmt.Sprintln("added csv col ", l, " content ", colToAdd)

		s.columns++
	}
	s.description = append(s.description, operation)

}

/*
 * Return a []string representing column c in s
 */
func (s *mCSVHolder) getCol(c int) []string {
	col := []string{}
	for _, row := range s.lines {
		col = append(col, row[c])
	}
	return col
}

/*
 * Return a deep copy of row r
 * TODO: Investigate why this isnt working
 */
func (s *mCSVHolder) getrRow(r int) []string {
	cpy := []string{}
	for _, str := range s.lines[r] {
		cpy = append(cpy, fmt.Sprintf("%s", str))
	}
	return cpy
}

/*
 * Insert a duplicate of column c
 */
func (s *mCSVHolder) copyCol(c int) {
	if c < s.columns {
		col := s.getCol(c)
		s.addColumn(c, col)

		s.description = append(s.description, fmt.Sprintln("coppied csv column", c))
	}
}

/*
 * Insert a duplicate of row r
 */
func (s *mCSVHolder) copyRow(r int) {
	if r < s.rows {
		c := s.getrRow(r)
		s.addRow(r, c)
		s.description = append(s.description, fmt.Sprintln("coppied csv row", r))
	}
}

/*
 * convert s into a string representing the content of the CSV
 */
func (s *mCSVHolder) flatten() string {
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
 * parse a string into the CSVHolder
 */
func (s *mCSVHolder) expand(in string) {
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
func (s *mCSVHolder) reExpand() {
	s.expand(s.flatten())
}

/*
 * Insert a blank row into position r
 */
func (s *mCSVHolder) addBlankRow(r int) {
	blankRow := make([]string, s.columns)
	s.addRow(r, blankRow)
}

/*
 * Insert a blank column into position c
 */
func (s *mCSVHolder) addBlankCol(c int) {
	blankCol := make([]string, s.rows)
	s.addColumn(c, blankCol)
}

/*
 * Deep copy (hopefully) s into a testcase
 */
func (s *mCSVHolder) generateTestCase() TestCase {
	ts := TestCase{}
	copy(ts.changes, s.description)
	content := s.flatten()
	copy(ts.input, content)

	return ts
}

/*
 * Teses making lots of copies of a row. If copies is true then copy the last row
 * otherwise add blank rows
 */
func spamRows(copies bool, tests chan<- TestCase, s mCSVHolder) {
	//TODO: abstract magic numbers
	for i := 1; i < 4096; i++ {
		// copy the last row
		if copies {
			s.copyRow(s.rows - 1)
		} else {
			s.addBlankRow(s.rows - 1)
		}

		//only send every nth case to test
		if i < 3 || i%16 == 0 {
			tests <- s.generateTestCase()
		}
	}
}

/*
 * Teses making lots of copies of a column. If copies is true then copy the last row
 * otherwise add blank rows
 */
 func spamCols(copies bool, tests chan<- TestCase, s mCSVHolder) {
	//TODO: abstract magic numbers
	for i := 1; i < 4096; i++ {
		// copy the last row
		if copies {
			s.copyCol(s.columns - 1)
		} else {
			s.addBlankCol(s.columns - 1)
		}

		//only send every nth case to test
		if i < 3 || i%16 == 0 {
			tests <- s.generateTestCase()
		}
	}
}

/*
 * Blank out the csv
 */
func blankCSV(copies bool, tests chan<- TestCase, s mCSVHolder) {

}

/*
 * Take a CSV file as base and permute variations into the test channel
 */
func generateCSVs(tests chan<- TestCase, file string) {
	input := newCSV("Initial input")
	input.read(file)

	//put the original input file into the test
	tests <- input.generateTestCase()

	//spam adding blank rows
	input = newCSV("Spamming blank CSV rows")
	input.read(file)
	spamRows(false, tests, input)

	//spam adding copies of the last row
	input = newCSV("Spamming copies CSV rows")
	input.read(file)
	spamRows(true, tests, input)

	//spamm adding blank columns
	input = newCSV("Spamming blank CSV cols")
	input.read(file)
	spamCols(false, tests, input)

	//spam adding copies of the last column
	input = newCSV("Spamming copies CSV cols")
	input.read(file)
	spamCols(true, tests, input)

	//TODO: probably close the channel here?
}

/*
 * visual test of the CSV generator
 */
func testCSVGenerator() {
	input := newCSV("Initial input")
	input.read("valid.csv")
	input.copyRow(0)
	input.copyRow(0)
	input.copyRow(0)
	input.copyRow(0)
	input.copyRow(0)
	input.reExpand()
	input.addBlankRow(0)
	input.copyRow(0)
	input.copyRow(0)
	input.reExpand()
	input.lines[0][0] = "s"
	input.lines[3][0] = "s"
	input.display()

}
