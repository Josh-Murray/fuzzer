package main

import (
	"fmt"
	"math"
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
	change += fmt.Sprint("Mutating some ints")
	return mutateObj(s1, isAInt, interestingInteger)
}

func isAFloat(w string) bool {

	_, err := strconv.ParseFloat(w, 1)

	if !isAInt(w) && err == nil {
		return true
	}
	return false
}

func interestingFloat(n int) string {
	candidates := []float32{0, -0, 1, -1, float32(math.Inf(1)), float32(math.Inf(1)), float32(math.NaN())}
	s1 := rand.NewSource(time.Now().UnixNano() * int64(n))
	r1 := rand.New(s1)
	result := candidates[r1.Intn(len(candidates))]
	return fmt.Sprintf("%v", result)
}

func mutateFloat(s1 string) string {
	change += fmt.Sprint("Mutating some floats")
	return mutateObj(s1, isAFloat, interestingFloat)
}

func mutateShuffle(s string) string {
	o := decompose(s)
	change += fmt.Sprint("Suffling")
	s1 := rand.NewSource(time.Now().UnixNano() * int64(1))
	r1 := rand.New(s1)
	r1.Shuffle(len(o), func(i, j int) { o[i], o[j] = o[j], o[i] })
	return compose(o)
}

func mutateReverse(s string) string {
	o := decompose(s)
	change += fmt.Sprint("Reversing entire thing")
	for left, right := 0, len(o)-1; left < right; left, right = left+1, right-1 {
		o[left], o[right] = o[right], o[left]
	}
	return compose(o)
}

func isAHex(w string) bool {
	_, err := strconv.ParseFloat(w, 1)

	if !isAFloat(w) && err == nil {
		return true
	}
	return false

}

func interestingHex(i int) string {
	candidates := []string{"0", "0x", "0x00000000", "0x0000000", "0xFFFFFFFF", "0x80000000", "0xdeadbeef", "01234567", "0xDEADBEEF", "0x0000000G"}
	s1 := rand.NewSource(time.Now().UnixNano() * int64(i))
	r1 := rand.New(s1)

	return candidates[r1.Intn(len(candidates))]
}

func mutateHex(s string) string {
	change += fmt.Sprint("Mutating some hex values")
	return mutateObj(s, isAHex, interestingHex)
}

/*
func main() {
	s2 := "This 22 is -22 a test  222 of 1.1 integers 333 222 333"
	o3 := mutateFloat(s2)
	fmt.Println(change)
	fmt.Println(o3)

}
*/
