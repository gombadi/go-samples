/*

This file contains code to decode the event given to a lambda function and return a
map[string]string with all the elements of the structure

I normally use node.js as the lambda function and provide the event to the Go code
as the last command line argument. It can be used as -

rc := GetLambdaRecordCount(os.Args[len(os.Args)-1])

*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// LambdaRecord contains a map[string]string of all elements extracted from a
// lambda event
type LambdaRecord struct {
	event map[string]string
}

// uncomment if you want to test compile the file
//func main() {}

// GetLambdaRecordCount returns the number of "Records" found in the event
func GetLambdaRecordCount(cmd string) int {

	lr := &lambdaRecord{}
	lr.event = make(map[string]string)

	var maprec map[string][]json.RawMessage

	err := json.Unmarshal([]byte(cmd), &maprec)
	if err != nil {
		return 0
	}
	return len(maprec["Records"])
}

// GetLambdaRecord returns details on the first "Records" found in the event
func GetLambdaRecord(cmd string) (*LambdaRecord, error) {
	// return details of the first record in the event
	return GetLambdaRecordIndex(cmd, 0)
}

// GetLambdaRecordIndex takes raw json supplied to the lambda function via node.js and returns
// a struct with a GetAttribute func to extract data
func GetLambdaRecordIndex(cmd string, index int) (*LambdaRecord, error) {

	lr := &LambdaRecord{}
	lr.event = make(map[string]string)

	var err error
	var maprec map[string][]json.RawMessage

	err = json.Unmarshal([]byte(cmd), &maprec)
	if err != nil {
		return lr, errors.New("unablr to find records in input")
	}

	if index > len(maprec["Records"]) {
		return lr, fmt.Errorf("record %d does not exist", index)
	}

	var event map[string]json.RawMessage
	err = json.Unmarshal(maprec["Records"][index], &event)
	if err != nil {
		return lr, fmt.Errorf("unable to extract event data from json records: %v", err)
	}

	for k, v := range event {
		if isJSON(v) {
			copyFromJSON(lr, k, v)
		} else {
			lr.event[strings.ToLower(k)] = string(v)
		}
	}
	return lr, nil
}

// GetAttribute returns the event attribute as a string or an empty string if not found
func (lr *lambdaRecord) GetAttribute(a string) string {
	s, _ := lr.GetAttributeBool(a)
	return s
}

// GetAttribute returns the event attribute as a string and true or an empty string and false
// if the attribute can not be found
func (lr *lambdaRecord) GetAttributeBool(a string) (string, bool) {

	// make sure the struct is valid before scanning it
	if lr == nil {
		return "", false
	}

	for k, v := range lr.event {
		if strings.EqualFold(a, k) {
			return cleanString(v), true
		}
	}
	return "", false
}

// PrintAttributes prints out a list of all known attributes. Debugging
func (lr *lambdaRecord) PrintAttributes() {
	for k, v := range lr.event {
		fmt.Printf("k: %s v: %s\n", k, v)
	}
}

// copyFromJSON recurcively copys elrments from the source json string
func copyFromJSON(lr *lambdaRecord, pk string, j json.RawMessage) {
	var tm map[string]json.RawMessage

	_ = json.Unmarshal(j, &tm)
	for k, v := range tm {
		if isJSON(v) {
			copyFromJSON(lr, pk+"."+k, v)
		} else {
			lr.event[strings.ToLower(pk+"."+k)] = string(v)
		}
	}
}

// isJSON returns true if it is given a json string
func isJSON(s json.RawMessage) bool {
	var js map[string]interface{}
	return json.Unmarshal(s, &js) == nil
}

func cleanString(s string) string {
	return strings.Trim(s, "\"")
}

// cleanString removes the " from the json.RawMessage and returns it as a string
func cleanjsonString(j json.RawMessage) string {
	return cleanString(string(j))
}

/*

 */
