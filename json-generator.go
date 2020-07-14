package main

import (
	"log"
	"strconv"
	"math/rand"
	"time"
)

// json value type enum
type valueType int
const (
	jString valueType = iota
	jInt
	jArray
	jObject
)

var (
	jInterestingStrings =  []string{"\"yeet\"", "\"swag\"", "\"aaaa\"", "\"bbbb\"", "\"0\"", "\"-1\"", "\"adam\"", "\"trivial\""}
	jInterestingInts = []int {-256, -128, -1, 0, 1, 16, 32, 64,127, 255, 256, 512, 1024, 4096, 0xffffffff, 0x7ffffffff}
)

type JsonElement struct {
	key string
	values []string
	vType valueType
}

type JsonHolder struct {
	rng *rand.Rand
	elements []JsonElement
}

func createJsonHolder(seed int64) JsonHolder{
	r := rand.New(rand.NewSource(seed))
	return JsonHolder{rng:r}
}

func (j * JsonHolder) addElement(newKey string, newValues []string, newType valueType) {
	newElem := JsonElement{key:newKey, values: make([]string, len(newValues)), vType: newType}
	copy(newElem.values, newValues)
	j.elements = append(j.elements, newElem)
}

func (j * JsonHolder) addInterestingString(keys []string, values []string){
	key := keys[j.rng.Intn(len(keys))]
	value := []string{keys[j.rng.Intn(len(keys))]}
	j.addElement(key, value, jString)
}

func (j * JsonHolder) addRandomInt(keys []string){
	key := keys[j.rng.Intn(len(keys))]
	value := []string{strconv.Itoa(j.rng.Intn(0xffffffff))}
	j.addElement(key, value, jInt)
}
// TODO add objects and arrays to this
func (j * JsonHolder) addArray(keys []string, ints []int){
	n := j.rng.Intn(0xffffffff)
	vals := []string{}
	key := keys[j.rng.Intn(len(keys))]
	for (n!=0){
		val:= (n%10)%2;
		if val == 0 {
			vals = append(vals, keys[j.rng.Intn(len(keys))])
		}else{
			vals = append(vals, strconv.Itoa(ints[j.rng.Intn(len(ints))]))
		}
		n /= 10
	}
	j.addElement(key, vals, jArray)
}

func (j * JsonHolder) flatten() string{
	result := "{\n"
	for n, i := range j.elements {
		result += "\t" + i.key + ": ";
		switch i.vType{
			case jArray:
				result +="[ "
				for m, j := range i.values {
					result += j
					if (m < len(i.values)-1){
						result += ", "
					}
				}
				result += "]"
			case jString:
				result += i.values[0]
			case jInt:
				result += i.values[0]

		}
		if (n < len(j.elements) -1 ){
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
					holder.addInterestingString(jInterestingStrings, jInterestingStrings)
				case 1:
					holder.addRandomInt(jInterestingStrings)
				case 2:
					holder.addArray(jInterestingStrings, jInterestingInts)
			}
		}
		holder.flatten()
	}
}
