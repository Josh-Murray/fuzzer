package main

import (
	"fmt"
	"os"
)

/*
 * Accepts an executable file and an input file to run it with.
 */
func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage:", os.Args[0], "<binary>", "<input file>")
	}

	//binary := os.Args[1]
	//inputFile := os.Args[2]

}
