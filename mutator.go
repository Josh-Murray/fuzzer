package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

/*
 * A Mutator has a reference to the bytes read in
 * from the input file provided to the fuzzer.
 * This is a temporary fix at the moment.
 * Mutators should be mutating TestCases received from a channel that
 * a permutator is sending to. Not all permutators have been implemented so
 * a mutator may not receive a TestCase from the channel inChan.
 * In this case, a mutator will create a new TestCase using fileBytes as
 * the input and perform mutate on that.
 */
type Mutator struct {
	// TODO setup channel for inputs and a worker function to use inputs
	/* TODO add configurables here
	 * rng seed
	 * num mutations
	 * etc
	 */
	inChan    chan TestCase
	outChan   chan TestCase
	rng       *rand.Rand
	fileBytes []byte
}

// TODO work out configurables, they might be needed here
func createMutator(out chan TestCase, in chan TestCase,
	fileBytes []byte, seed int64) Mutator {
	r := rand.New(rand.NewSource(seed))
	return Mutator{in, out, r, fileBytes}
}

func replace(o []string, changes *[]string, i int, v string) {
	c := fmt.Sprintf("Replacing input[%d]=%s with %s\n", i, o[i], v)
	*changes = append(*changes, c)
	o[i] = v
}

/*
 * splits the string s according to the regexp r using FindAllString.
 * FindAllString returns a slice of all successive matches.
 * The regexp currently finds matches for
 * 	- all ',' characters and all whitespace characters ([,\s])
 * 	- all other characters that are not ',' or whitespace ([^,\s]+)
 * splitting by both delimiters and everything else that are not said delimiter
 * means the resulting slice of strings includes the delimiter.
 * e.g. the string:
 * 	s := "1, 2, -100"
 * becomes the slice of strings
 * 	result := {"1", ",", " ", "2", ",", " ", "-100"}
 * This allows the decomposed string to be recomposed later on with the
 * delimiters intact.
 */
func decompose(s string) []string {
	r := regexp.MustCompile(`([,\s]|[^,\s]+)`)
	return r.FindAllString(s, -1)
}

func compose(o []string) string {
	return strings.Join(o, "")
}

/*
 * takes in a slice of strings o and identifies whether or not
 * each string in the slice can be parsed by the function s
 * (isAFloat, isAInt, isAHex).
 * returns a slice of indexes corresponding to the strings in o
 * which can successfully be parsed by the function s
 */
func identifyCandidates(o []string, s func(string) bool) []int {
	var candidates []int
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

func (m Mutator) interestingInteger(i int) string {
	candidates := []string{"0", "-1", "-100", "100", "4294967295", "-4294967295"}
	return candidates[m.rng.Intn(len(candidates))]
}

func (m Mutator) interestingString(i int) string {
	longString := strings.Repeat("a", 1024)
	candidates := []string{"%s%s%s%s%s", "%n%n%n%n%n", longString}
	return candidates[m.rng.Intn(len(candidates))]
}

func isAFloat(w string) bool {

	_, err := strconv.ParseFloat(w, 1)

	if !isAInt(w) && err == nil {
		return true
	}
	return false
}

func (m Mutator) interestingFloat(n int) string {
	candidates := []float32{0, -0, 1, -1, float32(math.Inf(1)), float32(math.Inf(1)), float32(math.NaN())}
	result := candidates[m.rng.Intn(len(candidates))]
	return fmt.Sprintf("%v", result)
}

/*
 * Determines whether or not a string is can be parsed as a number in
 * hex format. ParseInt accepts strings in 0x format however ParseFloat
 * does not.
 * If ParseInt does not throw an error but ParseFloat does, then
 * the string is a number in hex format.
 */
func isAHex(w string) bool {
	/* check if string starts with 0x or 0X, return if it doesn't */
	m, err := regexp.MatchString(`(m?)^0[xX]`, w)
	if err != nil || m == false {
		return false
	}

	_, err = strconv.ParseInt(w, 0, 64)
	if err != nil {
		return false
	}

	_, err = strconv.ParseFloat(w, 1)
	if err == nil {
		return false
	}

	return true
}

func (m Mutator) interestingHex(i int) string {
	candidates := []string{"0", "0x", "0x00000000", "0x0000000", "0xFFFFFFFF", "0x80000000", "0xdeadbeef", "01234567", "0xDEADBEEF", "0x0000000G"}
	return candidates[m.rng.Intn(len(candidates))]
}

func replaceN(s string, r *regexp.Regexp, n int, new string) string {

	locations := r.FindAllStringIndex(s, -1)
	result := ""

	if (n >= len(locations)) {
		n = 0
	}

	if n < len(locations) {
		start := locations[n][0]
		end := locations[n][1]

		result = fmt.Sprintf("%s%s%s", s[:start], new, s[end:])
		return result
	}
	return result

}

// Returns a slice of elements with exact count.
// step will be used as the increment between elements in the sequence.
// step should be given as a positive, negative or zero number.
// https://stackoverflow.com/questions/39868029/how-to-generate-a-sequence-of-numbers
// seriously go doesnt have something like this ?
func newSlice(start, count, step int) []int {
	s := make([]int, count)
	for i := range s {
		s[i] = start
		start += step
	}
	return s
}

/*
 * takes in a testCase which contains the input to mutate.
 * The input is decomposed into a slice of strings and the function cnd
 * determines which of the strings are suitable to apply the function
 * rplc on. The function rplc returns a string with which to replace some
 * value in the decomposed TestCase input 'o'.
 */
func (m Mutator) mutateObj(ts *TestCase, r *regexp.Regexp, rplc func(int) string) error {
	o := (string(ts.input))
	numbCandidates := len(r.FindAllStringIndex(o, -1))

	c := newSlice(0, numbCandidates, 1)

	/* return an error if there are no candidates to mutate */
	if len(c) == 0 {
		return errors.New("mutateObj: no candidates found")
	}

	/* shuffle the  candidate locations */
	m.rng.Shuffle(len(c), func(i, j int) { c[i], c[j] = c[j], c[i] })

	/* pick a random number of locations to change */
	numToChange := m.rng.Intn(len(c))
	changeLocs := c[:numToChange]
	for i, location := range changeLocs {
		/* replace the string at index location with the
		 * string returned from func rplc(i) */
		o = replaceN(o, r, location, rplc(i))

	}

	ts.input = []byte(o)

	return nil

}

func (m Mutator) mutateInts(ts *TestCase) error {
	regexInteger := regexp.MustCompile(`-?\d+`)
	result := m.mutateObj(ts, regexInteger, m.interestingInteger)
	return result
}

func (m Mutator) mutateFloats(ts *TestCase) error {
	regexFloat := regexp.MustCompile(`-?\d+\.\d+`)
	result := m.mutateObj(ts, regexFloat, m.interestingFloat)
	return result
}

func (m Mutator) mutateHex(ts *TestCase) error {
	regexHex := regexp.MustCompile(`(0x)?-?[0-9A-Fa-f]+`)
	result := m.mutateObj(ts, regexHex, m.interestingHex)
	return result
}

func (m Mutator) insertString(ts *TestCase) {
	insert := []byte(m.interestingString(0))
	where := m.rng.Intn(len(ts.input) + 1)
	if where == 0 {
		// prepend insert to the start of ts.input
		ts.input = append(insert, ts.input...)
	} else if where == len(ts.input) {
		// append insert to the end of ts.input
		ts.input = append(ts.input, insert...)
	} else {
		// add insert to the middle of ts.input
		head := ts.input[:where - 1]
		tail := ts.input[where:]
		ts.input = append(head, insert...)
		ts.input = append(ts.input, tail...)
	}

	return
}

func (m Mutator) mutateShuffle(ts *TestCase) {
	o := decompose(string(ts.input))
	m.rng.Shuffle(len(o), func(i, j int) { o[i], o[j] = o[j], o[i] })
	ts.changes = append(ts.changes, "Shuffling")
	ts.input = []byte(compose(o))
}

func (m Mutator) mutateReverse(ts *TestCase) {
	o := decompose(string(ts.input))
	for left, right := 0, len(o)-1; left < right; left, right = left+1, right-1 {
		o[left], o[right] = o[right], o[left]
	}
	ts.changes = append(ts.changes, "Reversing entire thing")
	ts.input = []byte(compose(o))

}

func (m Mutator) flipBits(ts *TestCase) {
	// cannot perform this on an empty slice
	if len(ts.input) == 0 {
		return
	}
	// flip bit in N% of bytes, could change to $config% bytes or randRange bytes
	size := float64(len(ts.input)) * 0.05
	nbytes := int(size)
	if nbytes == 0 {
		nbytes = 1
	}
	for i := 0; i < nbytes; i++ {
		index := m.rng.Intn(len(ts.input))
		offset := m.rng.Intn(8)
		msg := fmt.Sprintf("Mutator performed 'flip_bits' on byte %d, bit %d\n", index, offset)
		ts.changes = append(ts.changes, msg)
		ts.input[index] ^= 1 << offset
	}
}

func (m Mutator) flipBytes(ts *TestCase) {
	// cannot perform this on an empty slice
	if len(ts.input) == 0 {
		return
	}
	// flip N% of bytes
	size := float64(len(ts.input)) * 0.05
	nbytes := int(size)
	if nbytes == 0 {
		nbytes = 1
	}
	for i := 0; i < nbytes; i++ {
		index := m.rng.Intn(len(ts.input))

		msg := fmt.Sprintf("Mutator performed 'flip_bytes' on byte %d\n", index)
		ts.changes = append(ts.changes, msg)
		ts.input[index] ^= 255
	}
}

/* should monitor the probability this gets called to ensure the input
 * pool doesnt converge to 0 length inputs
 */
func (m Mutator) deleteSlice(ts *TestCase) {
	// used len too much, use a variable instead
	length := len(ts.input)
	if length == 0 {
		return
	}
	start := m.rng.Intn(length - 1)
	size := float64(length) * 0.2
	end := start + m.rng.Intn(int(size))
	if end > length {
		end = length
	}
	msg := fmt.Sprintf("Mutator performed 'delete_slice' on input[%d:%d]\n", start, end)
	ts.changes = append(ts.changes, msg)
	ts.input = append(ts.input[:start], ts.input[end:]...)
}

/* should monitor the probability this gets called to ensure the input
 * pool doesnt become too long
 */
func (m Mutator) duplicateSlice(ts *TestCase) {
	// used len too much, use a variable instead
	length := len(ts.input)

	if length == 0 {
		return
	}
	start := m.rng.Intn(length - 1)
	size := float64(length) * 0.2
	end := start + m.rng.Intn(int(size))
	if end > length {
		end = length
	}
	// this can probably be cleaned up
	tmp := make([]byte, len(ts.input[end:]))
	copy(tmp, ts.input[end:])
	msg := fmt.Sprintf("Mutator performed 'duplicate_slice' on '%s'\n", string(ts.input[start:end]))
	ts.changes = append(ts.changes, msg)
	ts.input = append(ts.input[:end], ts.input[start:end]...)
	ts.input = append(ts.input, tmp...)
}

// TODO add a int16 and int32 equivalent of this
func (m Mutator) interestingByte(ts *TestCase) {
	if len(ts.input) == 0 {
		return
	}
	interesting := []int8{-127, -1, 0, 1, 127, '{', '}', ',', '<', '>', '[', ']', '%'}
	val := interesting[m.rng.Intn(len(interesting))]
	pos := m.rng.Intn(len(ts.input))
	msg := fmt.Sprintf("Mutator performed 'interesting_byte' inserting int8 %d (%c) on byte %d\n", val, byte(val), pos)
	ts.changes = append(ts.changes, msg)
	ts.input[pos] = byte(val)
}

/*
 * returns a TestCase received from the mutators inChan channel. If no
 * input was received from the channel, a new TestCase is created
 * using the fileBytes stored in the mutator
 */
func (m Mutator) getTestCase() TestCase {
	var ts TestCase

	select {
	case ts = <-m.inChan:
	default:
		temp := append([]byte{}, m.fileBytes...)
		ts = TestCase{temp, []string{}}
	}

	return ts

}
func (m Mutator) mutateTestCase(ts TestCase) {
	nMutations := m.rng.Intn(8) + 1
	for i := 0; i < nMutations; i++ {
		selection := m.rng.Intn(8)
		// TODO work out configurables, they might be needed here
		switch selection {
		case 0:
			m.flipBits(&ts)
		case 1:
			m.flipBytes(&ts)
		case 2:
			err := m.mutateInts(&ts)
			if err != nil {
				continue
			}
		case 3:
			err := m.mutateFloats(&ts)
			if err != nil {
				continue
			}
		case 4:
			err := m.mutateHex(&ts)
			if err != nil {
				continue
			}
		case 5:
			m.mutateReverse(&ts)
		case 6:
			m.mutateShuffle(&ts)
		case 7:
			m.insertString(&ts)
		default:
			fmt.Printf("[WARN] mutator broken")
			//dunno
		}
	}
	m.outChan <- ts

}

// Same logic as mutate, but blocks when not receiving input
func (m Mutator) feedbackLoop() {
	for ts := range m.inChan {
		m.mutateTestCase(ts)
	}
}
func (m Mutator) mutate() {
	ts := m.getTestCase()
	m.mutateTestCase(ts)
}
