package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	o := []string{}
	g := []string{}
	fieldNames := []string{}
	fieldValues := []string{}
	f, err := os.Open("test.json")

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	s := bufio.NewScanner(f)
	for s.Scan() {
		//fmt.Println(s.Text())
		line := s.Text()
		if line == "{" {
			g = append(g, "o")
		} else if line == "}" {
			g = append(g, "c")
		} else {
			r := ""

			e := strings.Split(line, ":")
			if len(e) > 1 {
				r = fmt.Sprint("n:", e[0])
				fieldNames = append(fieldNames, e[0])
				if strings.Contains(e[1], "{") {
					r += fmt.Sprint("wo:", e[1])
				} else if strings.Contains(e[1], "[") {
					r += fmt.Sprint("so:", e[1])
				} else {
					r += fmt.Sprint(" v:", e[1])
					fieldValues = append(fieldValues, e[1])
				}
			} else {
				if strings.Contains(line, "}") {
					r += fmt.Sprint("cw:", line)
				} else if strings.Contains(line, "]") {
					r += fmt.Sprint("cs:", line)
				} else {
					r += fmt.Sprint("U*", line)
				}
			}
			g = append(g, r)
		}
		o = append(o, line)
	}
	err = s.Err()
	if err != nil {
		log.Fatal(err)
	}

	for i, line := range o {
		fmt.Println(i, g[i], "|||||", line)
	}
	fmt.Println(fieldNames)
	fmt.Println(fieldValues)
}
