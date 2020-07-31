package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
)

/* Interface that all json types implement. This allows grouped json elements
 * (arrays or objects) to just contain a generic type, and pass on the
 * responsiblity to the element to handle permutation/alterations.
 */
type jsonElement interface {
	flatten() string
	getKey() string
	spam(r *rand.Rand)
}

/// Json Strings
type jsonString struct {
	value string
	key   string
}

func (j *jsonString) flatten() string {
	return "\"" + j.value + "\""
}
func (j *jsonString) getKey() string {
	return "\"" + j.key + "\""
}

func (j *jsonString) spam(r *rand.Rand) {
	j.value = fmt.Sprintf("%s%s", j.value, j.value)
}

/// Json Ints
type jsonInt struct {
	value float64
	key   string
}

func (j *jsonInt) flatten() string {
	return fmt.Sprintf("%d", int(j.value))
}
func (j *jsonInt) getKey() string {
	return "\"" + j.key + "\""
}

func (j *jsonInt) spam(r *rand.Rand) {
	j.value += 0xffffffff
}

/// Json Array
type jsonArray struct {
	values []jsonElement
	key    string
}

func (j *jsonArray) flatten() string {
	result := "["
	for n, i := range j.values {
		result += i.flatten()
		if n < len(j.values)-1 {
			result += ", "
		}
	}
	result += "]"
	return result
}
func (j *jsonArray) getKey() string {
	return "\"" + j.key + "\""
}

func (j *jsonArray) spam(r *rand.Rand) {
	if len(j.values) == 0 {
		return
	}
	for i := 0; i < r.Intn(10)+3; i++ {
		// this is not a deep copy
		t := j.values[r.Intn(len(j.values))]
		j.values = append(j.values, t)
	}
}

/// Json Objects
type jsonObject struct {
	values []jsonElement
	key    string
}

func (j *jsonObject) flatten() string {
	result := "{"
	for n, i := range j.values {
		result += i.getKey() + ": " + i.flatten()
		if n < len(j.values)-1 {
			result += ", "
		}
	}
	result += "}"
	return result
}
func (j *jsonObject) getKey() string {
	return "\"" + j.key + "\""
}

func (j *jsonObject) spam(r *rand.Rand) {
	if len(j.values) == 0 {
		return
	}
	for i := 0; i < r.Intn(10)+3; i++ {
		// this is not a deep copy
		t := j.values[r.Intn(len(j.values))]
		j.values = append(j.values, t)
	}
}

/// Holder struct
type parsedJSON struct {
	toHarness chan TestCase
	toMutator chan TestCase
	rng       *rand.Rand
	json      *jsonCase
	original  map[string]interface{}
}

func newJSONPermutator(harness chan TestCase, mutator chan TestCase, seed int64) *parsedJSON {
	r := rand.New(rand.NewSource(seed))
	ip := &parsedJSON{toHarness: harness, toMutator: mutator, rng: r}
	return ip
}

/// Structure for the current permute case
type jsonCase struct {
	jsonObj []jsonElement
}

func (j *jsonCase) flatten() string {
	result := "{"
	for n, i := range j.jsonObj {
		result += i.getKey() + ": " + i.flatten()
		if n < len(j.jsonObj)-1 {
			result += ", "
		}
		result += ""
	}
	result += "}"
	return result
}

func (j *jsonCase) spam(r *rand.Rand) {
	if len(j.jsonObj) == 0 {
		return
	}
	for i := 0; i < r.Intn(10)+3; i++ {
		// this is not a deep copy
		t := j.jsonObj[r.Intn(len(j.jsonObj))]
		t.spam(r)
		j.jsonObj = append(j.jsonObj, t)
	}
}

/// Parsing input file functions
func expandArray(arr []interface{}) []jsonElement {
	var ret []jsonElement
	// Expand arrays into their structs, note array values have no keys
	for _, v := range arr {
		switch vv := v.(type) {
		case string:
			ret = append(ret, &jsonString{key: "", value: vv})
		case float64:
			ret = append(ret, &jsonInt{key: "", value: vv})
		case []interface{}:
			ret = append(ret, &jsonArray{key: "", values: expandArray(vv)})
		case map[string]interface{}:
			ret = append(ret, &jsonObject{key: "", values: expandObject(vv)})
		default:
			panic("Failed to parse json")
		}
	}
	return ret
}

func expandObject(obj map[string]interface{}) []jsonElement {
	var ret []jsonElement
	// Exapnd objects into their structs
	for k, v := range obj {
		switch vv := v.(type) {
		case string:
			ret = append(ret, &jsonString{key: k, value: vv})
		case float64:
			ret = append(ret, &jsonInt{key: k, value: vv})
		case []interface{}:
			ret = append(ret, &jsonArray{key: k, values: expandArray(vv)})
		case map[string]interface{}:
			ret = append(ret, &jsonObject{key: k, values: expandObject(vv)})
		default:
			panic("Failed to parse json")
		}
	}
	return ret
}

// Open file and set the original field in parsedJSON
func (i *parsedJSON) parseFile(fileName string) {
	data, err := ioutil.ReadFile(fileName)
	i.newCase()
	if err != nil {
		panic(err)
	}
	temp := make(map[string]interface{})
	err = json.Unmarshal(data, &temp)
	i.original = temp
}

func (i *parsedJSON) newCase() {
	i.json = &jsonCase{}
}

// Function to reset the current json case, permute a new testCase and push to chan
func (i *parsedJSON) generateTestCase() {
	i.newCase()
	i.json.jsonObj = expandObject(i.original)
	i.json.spam(i.rng)
	ts := TestCase{}
	content := i.json.flatten()
	ts.input = append(ts.input, content...)
	i.toMutator <- ts
	i.toHarness <- ts
}

// Main method for the permutor
func (i *parsedJSON) permuteInput(fileName string) {
	i.parseFile(fileName)
	// generate new cases forever
	for {
		i.generateTestCase()
	}
}

// Function for testing. Do not use this for the fuzzer
func testJPermutor(seed int64, fileName string) {
	harness := make(chan TestCase)
	mutator := make(chan TestCase)
	ip := newJSONPermutator(harness, mutator, seed)
	ip.parseFile(fileName)
	for i := 0; i < 10; i++ {
		ip.generateTestCase()
		log.Println("-----------------------------------------")
	}
}
