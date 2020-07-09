package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var change string

func replace(o []string, i int, v string) {
	change += fmt.Sprintln("Replacing:", i, " ", o[i], "with", v)
	o[i] = v
}

func decompose(s string) []string {
	return strings.Split(s, " ")
}

func compose(o []string) string {
	return strings.Join(o, " ")
}

func identifyCandidates(o []string, s func(string) bool) []int {
	candidates := make([]int, 1)
	for i, obj := range o {
		if s(obj) {
			candidates = append(candidates, i)
		}
	}
	return candidates
}

func isAInt(o string) bool {
	if _, err := strconv.Atoi(o); err == nil {
		return true
	}

	return false
}

func interestingInteger(i int) string {
	candidates := []string{"0", "-1", "-100", "100"}
	s1 := rand.NewSource(time.Now().UnixNano() * int64(i))
	r1 := rand.New(s1)

	return candidates[r1.Intn(len(candidates))]
}

func mutateObj(s string, cnd func(string) bool, rplc func(int) string) string {
	o := decompose(s)
	c := identifyCandidates(o, cnd)

	seed := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(seed)

	//shuffle the  candidate locations
	rand.Shuffle(len(c), func(i, j int) { c[i], c[j] = c[j], c[i] })

	//pick a random number of locations to change
	numToChange := rand.Intn(len(c))
	changeLocs := c[:numToChange]
	change += fmt.Sprintln("Replacing:", changeLocs)

	for i, location := range changeLocs {
		replace(o, location, rplc(i))
	}

	return compose(o)

}

func mutateInts(s1 string) string {
	return mutateObj(s1, isAInt, interestingInteger)
}

func main() {
	s2 := "This 22 is -22 a test  222 of 1.1 integers 333 222 333"
	o3 := mutateInts(s2)
	fmt.Println(change)
	fmt.Println(o3)

}
