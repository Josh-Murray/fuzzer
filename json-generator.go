package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"time"
)

// global slices of interesting values
var (
	interestingStrings = []string{"yeet", "swag", "aaaa", "bbbb", "0", "-1", "adam", "trivial"}
	interestingInts    = []int{-256, -128, -1, 0, 1, 16, 32, 64, 127, 255, 256, 512, 1024, 4096, 0xffffffff, 0x7ffffffff}
)

/* Interface for all Json entities (strings, numbers, arrays, objects)
 * Each entity will implement a flatten method which will return the entity
 * as a string
 */
type JsonElement interface {
	flatten() string
}

/// Json Strings
type JsonString struct {
	value string
}

func (j JsonString) flatten() string {
	return "\"" + j.value + "\""
}

/// Json Ints
type JsonInt struct {
	value int
}

func (j JsonInt) flatten() string {
	return strconv.Itoa(j.value)
}

/// Json Array
type JsonArray struct {
	values []JsonElement
}

func (j JsonArray) flatten() string {
	result := "["
	for i, val := range j.values {
		result += val.flatten()
		if i < len(j.values)-1 {
			result += ", "
		}
	}
	result += "]"
	return result
}

/// Json Objects
/* keys dont really matter to the generator, they are added only when a
 * jsonObject gets flattened
 */
type JsonObject struct {
	values []JsonElement
}

func (j JsonObject) flatten() string {
	result := "{"
	for n, i := range j.values {
		key := interestingStrings[n%len(interestingStrings)]
		result += "\"" + key + "\": " + i.flatten()
		if n < len(j.values)-1 {
			result += ","
		}
		result += ""
	}
	result += "}"
	return result
}

/// Json Holder
type JsonHolder struct {
	rng  *rand.Rand
	json JsonObject
}

func createJsonHolder(seed int64) JsonHolder {
	r := rand.New(rand.NewSource(seed))
	return JsonHolder{rng: r}
}

// adds an element to the root json object
func (j *JsonHolder) addElement(element JsonElement) {
	j.json.values = append(j.json.values, element)
}

// create a new JsonString and add it to the root object
func (j *JsonHolder) addInterestingString() {
	newValue := interestingStrings[j.rng.Intn(len(interestingStrings))]
	newString := JsonString{value: newValue}
	j.addElement(newString)
}

// create a new JsonInt and add it to the root object
func (j *JsonHolder) addRandomInt() {
	newValue := j.rng.Intn(0xffffffff)
	newInt := JsonInt{value: newValue}
	j.addElement(newInt)
}

// Create a new JsonArray and add it to the root object
func (j *JsonHolder) addArray() {
	newArr := JsonArray{}
	for i := 0; i < j.rng.Intn(8); i++ {
		// select a random object to add the new array
		switch j.rng.Intn(4) {
		case 0:
			newArr.values = append(newArr.values, JsonString{value: interestingStrings[j.rng.Intn(len(interestingStrings))]})
		case 1:
			newArr.values = append(newArr.values, JsonInt{value: interestingInts[j.rng.Intn(len(interestingInts))]})
		case 2:
			nestedArr := JsonArray{}
			j.addNestedArray(&nestedArr, 0)
			newArr.values = append(newArr.values, nestedArr)
		case 3:
			nestedObj := JsonObject{}
			j.addNestedObject(&nestedObj, 0)
			newArr.values = append(newArr.values, nestedObj)
		}
	}
	j.addElement(newArr)
}

// Create a new JsonObject and add it to the root object
func (j *JsonHolder) addObject() {
	newObj := JsonObject{}
	for i := 0; i < j.rng.Intn(8); i++ {
		// select a random object to add the new array
		switch j.rng.Intn(4) {
		case 0:
			newObj.values = append(newObj.values, JsonString{value: interestingStrings[j.rng.Intn(len(interestingStrings))]})
		case 1:
			newObj.values = append(newObj.values, JsonInt{value: interestingInts[j.rng.Intn(len(interestingInts))]})
		case 2:
			nestedArr := JsonArray{}
			j.addNestedArray(&nestedArr, 0)
			newObj.values = append(newObj.values, nestedArr)
		case 3:
			nestedObj := JsonObject{}
			j.addNestedObject(&nestedObj, 0)
			newObj.values = append(newObj.values, nestedObj)

		}
	}
	j.addElement(newObj)
}

// Handles adding elements to a nested array
func (j *JsonHolder) addNestedArray(jArray *JsonArray, depth int) {
	// limit depth to prevent infinite recursion
	if depth >= 2 {
		// add a bunch of ints or strings to the array
		for i := 0; i < j.rng.Intn(8); i++ {
			switch j.rng.Intn(2) {
			case 0:
				jArray.values = append(jArray.values, JsonString{value: interestingStrings[j.rng.Intn(len(interestingStrings))]})
			case 1:
				jArray.values = append(jArray.values, JsonInt{value: interestingInts[j.rng.Intn(len(interestingInts))]})
			}
		}
	} else {
		// add a bunch of random json elements to the array
		for i := 0; i < j.rng.Intn(8); i++ {
			switch j.rng.Intn(4) {
			case 0:
				jArray.values = append(jArray.values, JsonString{value: interestingStrings[j.rng.Intn(len(interestingStrings))]})
			case 1:
				jArray.values = append(jArray.values, JsonInt{value: interestingInts[j.rng.Intn(len(interestingInts))]})
			case 2:
				newArr := JsonArray{}
				j.addNestedArray(&newArr, depth+1)
				jArray.values = append(jArray.values, newArr)
			case 3:
				nestedObj := JsonObject{}
				j.addNestedObject(&nestedObj, 0)
				jArray.values = append(jArray.values, nestedObj)
			}
		}

	}
}

// Handles adding elements to a nested object
func (j *JsonHolder) addNestedObject(jObj *JsonObject, depth int) {
	// limit depth to prevent infinite recursion
	if depth >= 2 {
		// add a bunch of ints or strings to the array
		for i := 0; i < j.rng.Intn(8); i++ {
			switch j.rng.Intn(2) {
			case 0:
				jObj.values = append(jObj.values, JsonString{value: interestingStrings[j.rng.Intn(len(interestingStrings))]})
			case 1:
				jObj.values = append(jObj.values, JsonInt{value: interestingInts[j.rng.Intn(len(interestingInts))]})
			}
		}
	} else {
		for i := 0; i < j.rng.Intn(8); i++ {
			// add a bunch of random json elements to the array
			switch j.rng.Intn(3) {
			case 0:
				jObj.values = append(jObj.values, JsonString{value: interestingStrings[j.rng.Intn(len(interestingStrings))]})
			case 1:
				jObj.values = append(jObj.values, JsonInt{value: interestingInts[j.rng.Intn(len(interestingInts))]})
			case 2:
				newArr := JsonArray{}
				j.addNestedArray(&newArr, depth+1)
				jObj.values = append(jObj.values, newArr)
			case 3:
				newObj := JsonObject{}
				j.addNestedObject(&newObj, depth+1)
				jObj.values = append(jObj.values, newObj)
			}
		}

	}
}

// flatten the root json object
func (j *JsonHolder) flatten() []byte {
	result := j.json.flatten()
	var out bytes.Buffer
	err := json.Indent(&out, []byte(result), "", "  ")
	if err != nil {
		panic(err)
	}
	return out.Bytes()
}

// WARNING: this function is used only for testing purposes. Don't use this in the fuzzer
func testJGenerator() {
	for i := 0; i < 5; i++ {
		holder := createJsonHolder(time.Now().UnixNano())
		for i := 0; i < holder.rng.Intn(10)+1; i++ {
			choice := holder.rng.Intn(4)
			switch choice {
			case 0:
				holder.addInterestingString()
			case 1:
				holder.addRandomInt()
			case 2:
				holder.addArray()
			case 3:
				holder.addObject()
			}
		}
		validJson := holder.flatten()
		log.Println(string(validJson))
	}
}
