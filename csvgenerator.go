package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

type CSVHolder struct {
	lines         [][]string
	rows, columns int
	description   string
}

func newCSV(initialDescription string) CSVHolder {
	return CSVHolder{description: initialDescription}
}

func (s *CSVHolder) read(file string) {
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
}

func (s *CSVHolder) display() {
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

func (s *CSVHolder) deleteRow(u int) {
	if u < s.rows {
		s.lines = append(s.lines[:u], s.lines[u+1:]...)
		s.description += fmt.Sprintln("\nRemoved csv row ", u)
		s.rows--
	}

}

func (s *CSVHolder) deleteCol(u int) {

	if u < s.columns {
		for i, line := range s.lines {
			s.lines[i] = append(line[:u], line[u+1:]...)
		}
		s.description += fmt.Sprintln("\nRemoved csv row ", u)
		s.columns--
	}

}

func (s *CSVHolder) addRow(l int, rowToAdd []string) {
	if l < s.rows {
		if len(rowToAdd) == s.columns {
			end := s.lines[l:]
			beginning := append(s.lines[:l], rowToAdd)
			s.lines = append(beginning, end...)
			s.description += fmt.Sprintln("added csv row ", l, " content ", rowToAdd)
			s.rows++
		}
	}
}

func (s *CSVHolder) addColumn(l int, colToAdd []string) {
	if l > s.columns {
		l = s.columns
	}
	if len(colToAdd) == s.rows {
		for i := 0; i < s.rows; i++ {
			end := s.lines[i][l:]
			beginning := append(s.lines[i][:l], colToAdd[i])
			s.lines[i] = append(beginning, end...)

		}
		s.description += fmt.Sprintln("added csv col ", l, " content ", colToAdd)
		s.columns++
	}

}

func (s *CSVHolder) getCol(c int) []string {
	col := []string{}
	for _, row := range s.lines {
		col = append(col, row[c])
	}
	return col
}

func (s *CSVHolder) getrRow(r int) []string {
	cpy := []string{}
	for _, str := range s.lines[r] {
		cpy = append(cpy, fmt.Sprintf("%s", str))
	}
	return cpy
}

func (s *CSVHolder) copyCol(c int) {
	if c < s.columns {
		col := s.getCol(c)
		s.addColumn(c, col)
		s.description += fmt.Sprintln("coppied csv column", c)
	}
}

func (s *CSVHolder) copyRow(r int) {
	if r < s.rows {
		c := s.getrRow(r)
		s.addRow(r, c)
		s.description += fmt.Sprintln("coppied csv row", r)
	}
}

func (s *CSVHolder) flatten() string {
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

func (s *CSVHolder) expand(in string) {
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

func (s *CSVHolder) reExpand() {
	s.expand(s.flatten())
}

func (s *CSVHolder) addBlankRow(r int) {
	blankRow := make([]string, s.columns)
	s.addRow(r, blankRow)
}
func (s *CSVHolder) addBlankCol(r int) {
	blankCol := make([]string, s.rows)
	s.addColumn(r, blankCol)
}

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
