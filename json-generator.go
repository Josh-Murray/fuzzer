package main

import (
	"log"
	"strconv"
	"math/rand"
	"time"
)

type JsonElement interface {
	flatten() string
}
// json value type enum
type valueType int
const (
	jString valueType = iota
	jInt
	jArray
	jObject
)

var (
	jInterestingStrings =  []string{"yeet", "swag", "aaaa", "bbbb", "0", "-1", "adam", "trivial"}
	jInterestingInts = []int {-256, -128, -1, 0, 1, 16, 32, 64,127, 255, 256, 512, 1024, 4096, 0xffffffff, 0x7ffffffff}
)

func (j JsonString) flatten() string{
	return  "\"" + j.value + "\""
}
type JsonString struct {
	value string
}

func (j JsonInt) flatten() string{
	return strconv.Itoa(j.value)
}
type JsonInt struct {
	value int
}
func (j JsonArray) flatten() string{
	result := "["
	for i, val := range j.values {
		result += val.flatten()
		if i < len(j.values) -1 {
			result += ", "
		}
	}
	result += "]"
	return result
}
type JsonArray struct {
	values []JsonElement
}

type JsonObject struct {
	keys []string
	values []JsonElement
}

type JsonHolder struct {
	rng *rand.Rand
	json JsonObject
}

func createJsonHolder(seed int64) JsonHolder{
	r := rand.New(rand.NewSource(seed))
	return JsonHolder{rng:r}
}

func (j * JsonHolder) addElement(element JsonElement) {
	j.json.values = append(j.json.values, element)
}

func (j * JsonHolder) addInterestingString(strs []string){
	newValue := strs[j.rng.Intn(len(strs))]
	newString := JsonString{value: newValue}
	j.addElement(newString)
}

func (j * JsonHolder) addRandomInt(){
	newValue := j.rng.Intn(0xffffffff)
	newInt := JsonInt{value: newValue}
	j.addElement(newInt)
}


func (j * JsonHolder) addNestedArray(jArray * JsonArray, strs []string, ints []int, depth int){
	if depth == 3 {
		for i := 0; i < j.rng.Intn(8); i ++{
			switch j.rng.Intn(2){
				case 0:
					jArray.values = append(jArray.values, JsonString{value: strs[j.rng.Intn(len(strs))]})
				case 1:
					jArray.values = append(jArray.values, JsonInt{value: ints[j.rng.Intn(len(ints))]})
			}
		}
	}else{
		for i := 0; i < j.rng.Intn(8); i ++{
			switch j.rng.Intn(3){
				case 0:
					jArray.values = append(jArray.values, JsonString{value: strs[j.rng.Intn(len(strs))]})
				case 1:
					jArray.values = append(jArray.values, JsonInt{value: ints[j.rng.Intn(len(ints))]})
				case 2:
					newArr := JsonArray{}
					j.addNestedArray(&newArr, strs, ints, depth + 1)
					jArray.values = append(jArray.values, newArr)
			}
		}

	}
}
// TODO add objects and arrays to this
func (j * JsonHolder) addArray(strs []string, ints []int){
	newArr := JsonArray{}
	for i := 0; i < j.rng.Intn(8); i++{
		switch (j.rng.Intn(3)){
			case 0:
				newArr.values = append(newArr.values, JsonString{value: strs[j.rng.Intn(len(strs))]})
			case 1:
				newArr.values = append(newArr.values, JsonInt{value: ints[j.rng.Intn(len(ints))]})
			case 2:
				nestedArr := JsonArray{}
				j.addNestedArray(&nestedArr,strs, ints, 0)
				newArr.values = append(newArr.values, nestedArr)
		}
	}
	j.addElement(newArr)
}

func (j * JsonHolder) flatten(strs []string, ints []int) string{
	result := "{\n"
	for n, i := range j.json.values {
		key := strs[j.rng.Intn(len(strs))]
		result += "\t\"" + key + "\": " + i.flatten();
		if (n < len(j.json.values) -1 ){
			result += ","
		}
		result += "\n"
	}
	result += "}"
	log.Println(result)
	return result
}

func testJGenerator(){
	for i:=0; i < 5; i++ {
		holder := createJsonHolder(time.Now().UnixNano())
		for i:= 0; i < holder.rng.Intn(10) +1; i++ {
			choice := holder.rng.Intn(3)
			switch(choice){
				case 0:
					holder.addInterestingString(jInterestingStrings)
				case 1:
					holder.addRandomInt()
				case 2:
					holder.addArray(jInterestingStrings, jInterestingInts)
			}
		}
		holder.flatten(jInterestingStrings, jInterestingInts)
	}
}
