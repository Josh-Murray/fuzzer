package main

import (
	"math/rand"
	"encoding/json"
	"log"
	"io/ioutil"
	"fmt"
)
type jsonElement interface {
	flatten() string
	getKey() string
	spam(r *rand.Rand)
}

/// Json Strings
type jsonString struct {
	value string
	key string
}

func (j *jsonString) flatten() string {
	return  "\"" + j.value + "\""
}
func (j *jsonString) getKey() string {
	return "\"" + j.key+"\""
}

func (j *jsonString) spam(r *rand.Rand) {
	j.value = fmt.Sprintf("%s%s", j.value, j.value)
}
/// Json Ints
type jsonInt struct {
	value float64
	key string
}

func (j *jsonInt) flatten() string {
	return fmt.Sprintf("%d", int(j.value))
}
func (j *jsonInt) getKey() string {
	return "\"" + j.key+"\""
}

func (j *jsonInt) spam(r * rand.Rand) {
	j.value += 0xffffffff
}
/// Json Array
type jsonArray struct {
	values []jsonElement
	key string
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
	return "\"" + j.key+"\""
}

func (j *jsonArray) spam(r * rand.Rand){
	if len(j.values) == 0{
		return
	}
	for i:= 0; i < r.Intn(10)+3; i++ {
		// this is not a deep copy
		t := j.values[r.Intn(len(j.values))]
		j.values = append(j.values, t)
	}
}
/// Json Objects
type jsonObject struct {
	values []jsonElement
	key string
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
	return "\"" + j.key+"\""
}

func (j *jsonObject) spam(r * rand.Rand){
	if len(j.values) == 0{
		return
	}
	for i:= 0; i < r.Intn(10)+3; i++ {
		// this is not a deep copy
		t := j.values[r.Intn(len(j.values))]
		j.values = append(j.values, t)
	}
}

/// Holder struct
type parsedJSON struct {
	rng *rand.Rand
	json *jsonCase
}
/// Structure for the current case
type jsonCase struct {
	jsonObj []jsonElement
	description []string
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

func (j* jsonCase) spam(r * rand.Rand) {
	for i:= 0; i < r.Intn(10) + 3; i++ {
		// this is not a deep copy
		t := j.jsonObj[r.Intn(len(j.jsonObj))]
		t.spam(r)
		j.jsonObj = append(j.jsonObj, t)
	}
}

/// Parsing input file functions
func expandArray(arr []interface{}) []jsonElement {
	var ret []jsonElement
	for  _,v := range arr {
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

func (i *parsedJSON) parseFile(fileName string) {
	data, err := ioutil.ReadFile(fileName)
	i.newCase()
	if err != nil {
		panic(err)
	}
	temp := make(map[string]interface{})
	err = json.Unmarshal(data, &temp)
	i.json.jsonObj = expandObject(temp)

}
func (i * parsedJSON) newCase(){
	i.json = &jsonCase{}
}
func testJPermutor(seed int64, fileName string){
	r := rand.New(rand.NewSource(1))
	ip := &parsedJSON{rng: r}
	ip.parseFile(fileName)
	log.Println("============ BEFORE ================")
	log.Println(ip.json.flatten())
	ip.json.spam(ip.rng)
	log.Println("============ AFTER ================")
	log.Println(ip.json.flatten())
}
