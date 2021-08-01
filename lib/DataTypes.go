package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// type DataNotes struct {
// 	title string
// 	link  string
// 	note  string
// }

// type DataHints struct {
// 	userId     string
// 	pre        string
// 	post       string
// 	link       string
// 	positional string
// 	notes      string
// }

// type DataGroup struct {
// 	owner string
// 	notes []DataNotes
// 	hints []DataHints
// }

var (
	tabdata = "                                     "
)

type DataRoot struct {
	timeStamp string
	data      map[string]interface{}
}

func NewDataRoot(m map[string]interface{}) (*DataRoot, error) {
	ts, ok := m["timeStamp"]
	if !ok {
		return nil, errors.New("'timeStamp' does not exist in data root")
	}
	if reflect.ValueOf(ts).Kind() != reflect.String {
		return nil, errors.New("'timeStamp' is not a string")
	}
	_, ok = m["groups"]
	if !ok {
		return nil, errors.New("'groups' does not exist in data root")
	}
	return &DataRoot{timeStamp: fmt.Sprintf("%v", ts), data: m}, nil
}

func Parse(j []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(j), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *DataRoot) ToJson() (string, error) {
	output, err := json.MarshalIndent(r.data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (r *DataRoot) ToStruct() string {
	var sb strings.Builder
	printMapStruct(&sb, r.data, 1)
	return sb.String()
}

func printMapStruct(sb *strings.Builder, m map[string]interface{}, ind int) {
	for k, v := range m {
		if reflect.ValueOf(v).Kind() != reflect.String {
			sb.WriteString(fmt.Sprintf("%d:%s:%s \n", ind, tabdata[:ind*2], k))
			printMapStruct(sb, v.(map[string]interface{}), ind+1)
		} else {
			sb.WriteString(fmt.Sprintf("%d:%s-%s = %s\n", ind, tabdata[:ind*2], k, v))
		}
	}

}
