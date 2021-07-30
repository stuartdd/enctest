package lib

import (
	"encoding/json"
)

type DataNotes struct {
	title string
	link  string
	note  string
}

type DataHints struct {
	userId     string
	pre        string
	post       string
	link       string
	positional string
	notes      string
}

type DataGroup struct {
	owner string
	notes []DataNotes
	hints []DataHints
}

type DataRoot struct {
	timeStamp string
	groups    []DataGroup
}

func Parse(j []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(j), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func PrettyJson(m map[string]interface{}) (string, error) {
	output, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}
