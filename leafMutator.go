package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

//func (m Mutator) swapSymbol(ts *TestCase) {
//pick a symbol and swap it with a symbol of the same type
//assume seperator is ' ' but might either detect or include what the seperator is in TestCase
//	input := string(ts.input)
//	result, msg := smartReplaceRandom(input)

//	ts.changes = append(ts.changes, msg)
//	ts.input = []byte(result)

//}

func smartReplace(word string) (string, string) {
	var t, v string

	if _, err := strconv.Atoi(word); err == nil {
		t = "integer"
		v = fmt.Sprintf("%v", rand.Uint32())
	} else if _, err := strconv.ParseFloat(word, 1); err == nil {
		t = "float"
		v = fmt.Sprintf("%v", rand.ExpFloat64())
	} else if _, err := strconv.ParseInt(word, 0, 64); err == nil {
		t = "hex"
		// TODO: check for octal
		v = string("0xdeadbeef")
	} else {
		t = "word"
		v = string("R4nd0m")
	}

	return v, t
}

func smartReplaceRandom(original string) (string, string) {
	words := strings.Split(original, " ")

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	changeLocation := r1.Intn(len(words))

	oldWord := words[changeLocation]
	newWord, t := smartReplace(words[changeLocation])
	words[changeLocation] = newWord
	newString := strings.Join(words, " ")

	fmt.Println("Original String:", original)
	fmt.Println("Changing location:", changeLocation, "type:", t, "from:", oldWord, "to:", newWord)
	fmt.Println("New String:", newString)

	change := fmt.Sprintln("Changing location:", changeLocation, "type:", t, "from:", oldWord, "to:", newWord)

	return newString, change
}

func interestingInt() int {
	candidates := []int{0, -1, -100, 100}
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	return candidates[r1.Intn(len(candidates))]
}

func replaceRandInts(original string) (string, string) {
	words := strings.Split(original, " ")
	locations := make([]int, 0)
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	var change string

	//find location of all integers in the input
	for i, word := range words {
		if _, err := strconv.Atoi(word); err == nil {
			locations = append(locations, i)
		}
	}

	//shuffle the locations
	r1.Shuffle(len(locations), func(i, j int) { locations[i], locations[j] = locations[j], locations[i] })

	//pick a random number of locations to change
	numToChange := r1.Intn(len(locations))
	changeLocs := locations[:numToChange]
	change = fmt.Sprintln("Replacing ints:", changeLocs)
	for i, location := range changeLocs {
		oldVal := words[location]
		newVal := fmt.Sprintf("%v", interestingInt())
		words[location] = newVal
		change += fmt.Sprintln("replaceRandInt", i, " - location:", location, "from:", oldVal, "to: ", newVal)
	}

	newString := strings.Join(words, " ")
	return newString, change
}

func main() {
	var myInput string = "This -1 is 0x10f test -1.23 1090 input 1 10 1.30 with funn3y characters"
	fmt.Println("Original String:", myInput)
	//smartReplaceRandom(myInput)
	result, msg := replaceRandInts(myInput)
	fmt.Print(msg)
	fmt.Println(result)
}
