package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

var (
	tabdata = "                                     "
)

type DataRoot struct {
	timeStamp time.Time
	dataMap   map[string]interface{}
	navIndex  map[string][]string
}

func NewDataRoot(j []byte) (*DataRoot, error) {
	m, err := parseJson(j)
	if err != nil {
		return nil, err
	}
	ts, ok := m["timeStamp"]
	if !ok {
		return nil, errors.New("'timeStamp' does not exist in data root")
	}
	if reflect.ValueOf(ts).Kind() != reflect.String {
		return nil, errors.New("'timeStamp' is not a string")
	}
	tim, err := parseTime(fmt.Sprintf("%v", ts))
	if err != nil {
		return nil, errors.New("'timeStamp' could not be parsed")
	}
	_, ok = m["groups"]
	if !ok {
		return nil, errors.New("'groups' does not exist in data root")
	}

	dr := &DataRoot{timeStamp: tim, dataMap: m, navIndex: createNavIndex(m)}
	return dr, nil
}

func (r *DataRoot) GetDataRootMap() *map[string]interface{} {
	return &r.dataMap
}

func (r *DataRoot) GetTimeStamp() time.Time {
	return r.timeStamp
}

func (r *DataRoot) GetNavIndex(id string) []string {
	return r.navIndex[id]
}

func (r *DataRoot) ToJson() (string, error) {
	output, err := json.MarshalIndent(r.dataMap, "", "    ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (r *DataRoot) ToStruct() string {
	var sb strings.Builder
	appendMapStruct(&sb, r.dataMap, 1)
	return sb.String()
}

func appendMapStruct(sb *strings.Builder, m map[string]interface{}, ind int) {
	for k, v := range m {
		if reflect.ValueOf(v).Kind() != reflect.String {
			sb.WriteString(fmt.Sprintf("%d:%s:%s \n", ind, tabdata[:ind*2], k))
			appendMapStruct(sb, v.(map[string]interface{}), ind+1)
		} else {
			sb.WriteString(fmt.Sprintf("%d:%s-%s = %s\n", ind, tabdata[:ind*2], k, v))
		}
	}
}

func createNavIndex(m map[string]interface{}) map[string][]string {
	var uids = make(map[string][]string)
	createNavIndexDetail("", uids, m)
	return uids
}

func createNavIndexDetail(id string, uids map[string][]string, m map[string]interface{}) {
	for _, v := range m {
		if reflect.ValueOf(v).Kind() != reflect.String {
			l, ll := keysToList(id, v.(map[string]interface{}))
			uids[id] = ll
			for ii2, id2 := range l {
				m2 := v.(map[string]interface{})[id2]
				for _, v2 := range m2.(map[string]interface{}) {
					if reflect.ValueOf(v2).Kind() != reflect.String {
						l2, ll2 := keysToList(ll[ii2], m2.(map[string]interface{}))
						uids[ll[ii2]] = ll2
						for ii3, id3 := range l2 {
							m3 := m2.(map[string]interface{})[id3]
							for _, v3 := range m3.(map[string]interface{}) {
								if reflect.ValueOf(v3).Kind() != reflect.String {
									_, ll3 := keysToList(ll2[ii3], m3.(map[string]interface{}))
									uids[ll2[ii3]] = ll3
								}
								// else {
								// 	_, ll4 := keysToList(ll2[ii3], m3.(map[string]interface{}))
								// 	uids[ll2[ii3]] = ll4
								// }
							}
						}
					}
				}
			}
		}
	}
}

func keysToList(id string, m map[string]interface{}) ([]string, []string) {
	l := make([]string, 0)
	ll := make([]string, 0)
	for k, _ := range m {
		l = append(l, k)
		if id == "" {
			ll = append(ll, fmt.Sprintf("%s", k))
		} else {
			ll = append(ll, fmt.Sprintf("%s.%s", id, k))
		}

	}
	sort.Strings(l)
	sort.Strings(ll)
	return l, ll
}

func parseTime(st string) (time.Time, error) {
	t, err := time.Parse(time.UnixDate, st)
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}

func parseJson(j []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(j), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
