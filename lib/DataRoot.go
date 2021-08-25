package lib

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	hintStr      = "pwHint"
	noteStr      = "notes"
	groupsStr    = "groups"
	timeStampStr = "timeStamp"
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
	ts, ok := m[timeStampStr]
	if !ok {
		return nil, fmt.Errorf("'%s' does not exist in data root", timeStampStr)
	}
	if reflect.ValueOf(ts).Kind() != reflect.String {
		return nil, fmt.Errorf("'%s' is not a string", timeStampStr)
	}
	tim, err := parseTime(fmt.Sprintf("%v", ts))
	if err != nil {
		return nil, fmt.Errorf("'%s' could not be parsed", timeStampStr)
	}
	_, ok = m[groupsStr]
	if !ok {
		return nil, fmt.Errorf("'%s' does not exist in data root", groupsStr)
	}

	dr := &DataRoot{timeStamp: tim, dataMap: m, navIndex: createNavIndex(m)}
	return dr, nil
}

func (p *DataRoot) Search(addPath func(string), needle string) {
	rootMap := p.dataMap
	groups := rootMap[groupsStr].(map[string]interface{})
	for user, v := range groups {
		searchUsers(addPath, needle, user, v.(map[string]interface{}))
	}
}

func searchUsers(addPath func(string), needle, user string, m map[string]interface{}) {
	for k, v := range m {
		if k == hintStr {
			for k1, v1 := range v.(map[string]interface{}) {
				searchLeafNodes(addPath, hintStr, needle, user, k1, v1.(map[string]interface{}))
			}
		} else {
			searchLeafNodes(addPath, "", needle, user, k, v.(map[string]interface{}))
		}
	}
}

func searchLeafNodes(addPath func(string), tag, needle, user, name string, m map[string]interface{}) {
	if tag != "" {
		tag = "." + tag + "."
	} else {
		tag = "."
	}
	if strings.Contains(name, needle) {
		addPath(user + tag + name)
	}
	for k, s := range m {
		if strings.Contains(k, needle) {
			addPath(user + tag + name)
		} else {
			if reflect.ValueOf(s).Kind() == reflect.String {
				if strings.Contains(s.(string), needle) {
					addPath(user + tag + name)
				}
			} else {
				searchLeafNodes(addPath, name, needle, user, k, s.(map[string]interface{}))
			}
		}
	}
}

func (p *DataRoot) AddUser(userName string) (string, error) {
	m := GetMapForUid(userName, &p.dataMap)
	if m != nil {
		return "", fmt.Errorf("user name '%s' already exists", userName)
	}
	rootMap := p.dataMap
	groups := rootMap[groupsStr].(map[string]interface{})

	newUser := make(map[string]interface{})
	addHint("application", userName, newUser)
	addNote(userName, "note", "note for user "+userName, newUser)

	groups[userName] = newUser
	p.navIndex = createNavIndex(p.dataMap)
	return userName, nil
}

func (p *DataRoot) AddNote(userName, noteName, content string) (string, error) {
	user := GetMapForUid(userName, &p.dataMap)
	if user == nil {
		return "", fmt.Errorf("user name '%s' does not exists", userName)
	}
	defer func() {
		p.navIndex = createNavIndex(p.dataMap)
	}()
	return addNote(userName, noteName, content, *user)
}

func (p *DataRoot) AddHint(userName, appName string) (string, error) {
	user := GetMapForUid(userName, &p.dataMap)
	if user == nil {
		return "", fmt.Errorf("user name '%s' does not exists", userName)
	}
	defer func() {
		p.navIndex = createNavIndex(p.dataMap)
	}()
	return addHint(appName, userName, *user)
}

func addNote(userName, noteName, content string, user map[string]interface{}) (string, error) {
	_, ok := user[noteStr]
	if !ok {
		user[noteStr] = make(map[string]interface{})
	}
	notes := user[noteStr].(map[string]interface{})

	_, ok = notes[noteName]
	if !ok {
		notes[noteName] = content
		return fmt.Sprintf("%s.%s", userName, noteStr), nil
	}
	return "", fmt.Errorf("note '%s' already exists", noteName)
}

func addHint(hintName, userName string, user map[string]interface{}) (string, error) {
	_, ok := user[hintStr]
	if !ok {
		user[hintStr] = make(map[string]interface{})
	}
	pwHints := user[hintStr].(map[string]interface{})
	_, ok = pwHints[hintName]
	if !ok {
		pwHints[hintName] = make(map[string]interface{})
		hint := pwHints[hintName].(map[string]interface{})
		hint["userId"] = ""
		hint["pre"] = ""
		hint["post"] = ""
		hint["notes"] = ""
		hint["positional"] = "12345"
		return fmt.Sprintf("%s.%s.%s", userName, hintStr, hintName), nil
	}
	return "", fmt.Errorf("application '%s' already exists", hintName)
}
func GetParentId(uid string) string {
	if uid == "" {
		return uid
	}
	p := strings.LastIndexByte(uid, '.')
	switch p {
	case -1:
		return uid
	case 0:
		return ""
	default:
		return uid[0:p]
	}
}

func (p *DataRoot) GetRootUidOrCurrent(current string) string {
	if current != "" {
		return current
	}
	l := make([]string, 0)
	for k := range p.navIndex {
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
	x := n[groupsStr]
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
	for k := range m {
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
