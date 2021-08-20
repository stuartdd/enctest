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

func (p *DataRoot) AddUser(userName string) (string, error) {
	m := GetMapForUid(userName, &p.dataMap)
	if m != nil {
		return "", fmt.Errorf("user name '%s' already exists", userName)
	}
	rootMap := p.dataMap
	groups := rootMap["groups"].(map[string]interface{})

	newUser := make(map[string]interface{})
	addApplication("application", newUser)
	addNote(userName, "note", "note for user "+userName, newUser)

	groups[userName] = newUser
	p.navIndex = createNavIndex(p.dataMap)
	return userName, nil
}

func (p *DataRoot) AddNote(userName, noteName, content string) (string, error) {
	user := GetMapForUid(userName, &p.dataMap)
	if user == nil {
		return "", fmt.Errorf("user name '%s' doen not exists", userName)
	}
	return addNote(userName, noteName, content, *user)
}

func (p *DataRoot) AddApplication(userName, appName string) error {
	user := GetMapForUid(userName, &p.dataMap)
	if user == nil {
		return fmt.Errorf("user name '%s' doen not exists", userName)
	}
	return addApplication(appName, *user)
}

func addNote(userName, noteName, content string, user map[string]interface{}) (string, error) {
	_, ok := user["notes"]
	if !ok {
		user["notes"] = make(map[string]interface{})
	}
	notes := user["notes"].(map[string]interface{})

	_, ok = notes[noteName]
	if !ok {
		notes[noteName] = content
		return fmt.Sprintf("%s.notes", userName), nil
	}
	return "", fmt.Errorf("note '%s' already exists", noteName)
}

func addApplication(appName string, user map[string]interface{}) error {
	_, ok := user["pwHints"]
	if !ok {
		user["pwHints"] = make(map[string]interface{})
	}
	pwHints := user["pwHints"].(map[string]interface{})
	_, ok = pwHints[appName]
	if !ok {
		pwHints[appName] = make(map[string]interface{})
		application := pwHints[appName].(map[string]interface{})
		application["userId"] = ""
		application["pre"] = ""
		application["post"] = ""
		application["notes"] = ""
		application["positional"] = "12345"
		return nil
	}
	return fmt.Errorf("application '%s' already exists", appName)
}

func (p *DataRoot) GetRootUid(current string) string {
	if current != "" {
		return current
	}
	l := make([]string, 0)
	for k, _ := range p.navIndex {
		if k != "" {
			l = append(l, k)
		}
	}

	if len(l) > 0 {
		sort.Strings(l)
		return l[0]
	}
	return ""
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

func (r *DataRoot) ToJsonTreeMap() string {
	output, err := json.MarshalIndent(r.navIndex, "", "    ")
	if err != nil {
		return err.Error()
	}
	return string(output)
}

func (r *DataRoot) ToStruct() string {
	var sb strings.Builder
	appendMapStruct(&sb, r.dataMap, 1)
	return sb.String()
}

func GetMapForUid(uid string, m *map[string]interface{}) *map[string]interface{} {
	nodes := strings.Split(uid, ".")
	n := *m
	x := n["groups"]
	for _, v := range nodes {
		y := x.(map[string]interface{})[v]
		if y == nil {
			return nil
		}
		if reflect.ValueOf(y).Kind() != reflect.String {
			x = y
		}
	}
	o := x.(map[string]interface{})
	return &o
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
			ll = append(ll, k)
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
