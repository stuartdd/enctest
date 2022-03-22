/*
 * Copyright (C) 2021 Stuart Davies (stuartdd)
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
package lib

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/stuartdd2/JsonParser4go/parser"
)

type NodeAnnotationEnum int

const (
	IdHints            = "pwHints"
	IdNotes            = "notes"
	dataMapRootName    = "groups"
	timeStampName      = "timeStamp"
	tabdata            = "                                     "
	allowedCharsInName = " *@#$%^&*()_+=?"
	dateTimeFormatStr  = "Mon Jan 2 15:04:05 MST 2006"
	PATH_SEP           = "|"
	PATH_SEP_CHAR      = '|'

	NOTE_TYPE_SL NodeAnnotationEnum = 0 // Single Line: These are indexes. Found issues when using iota!
	NOTE_TYPE_ML NodeAnnotationEnum = 1 // Multi Line
	NOTE_TYPE_RT NodeAnnotationEnum = 2 // Rich Text
	NOTE_TYPE_PO NodeAnnotationEnum = 3 // POsitinal
	NOTE_TYPE_IM NodeAnnotationEnum = 4 // IMage
)

var (
	nodeAnnotationPrefix      = []string{"", "!ml", "!rt", "!po", "!im"}
	NodeAnnotationPrefixNames = []string{"Single Line", "Multi Line", "Rich Text", "Positional", "Image"}
	NodeAnnotationEnums       = []NodeAnnotationEnum{NOTE_TYPE_SL, NOTE_TYPE_ML, NOTE_TYPE_RT, NOTE_TYPE_PO, NOTE_TYPE_IM}
	NodeAnnotationsSingleLine = []bool{true, false, false, true, true}
	defaultHintNames          = []string{"notes", "post", "pre", "userId"}
	defaultAssetNames         = []string{"Account Num.", "Sort Code"}
	timeStampPath             = parser.NewBarPath(timeStampName)
	dataMapRootPath           = parser.NewBarPath(dataMapRootName)
)

type JsonData struct {
	dataMap        *parser.JsonObject
	navIndex       *map[string][]string
	dataMapUpdated func(string, *parser.Path, error)
}

func NewJsonData(j []byte, dataMapUpdated func(string, *parser.Path, error)) (*JsonData, error) {
	mIn, err := parser.Parse(j)
	if err != nil {
		return nil, err
	}
	if mIn.GetNodeType() != parser.NT_OBJECT {
		return nil, fmt.Errorf("root element is NOT a JsonObject")
	}
	rO := mIn.(*parser.JsonObject)

	u := rO.GetNodeWithName(dataMapRootName)
	if u == nil {
		return nil, fmt.Errorf("root '%s' element is missing", dataMapRootName)
	}
	if u.GetNodeType() != parser.NT_OBJECT {
		return nil, fmt.Errorf("root '%s' element is NOT a JsonObject", dataMapRootName)
	}
	ts, err := parser.Find(rO, timeStampPath)
	if err != nil {
		return nil, fmt.Errorf("'%s' does not exist in data root", timeStampPath)
	}
	_, err = parseTime((ts.(*parser.JsonString)).GetValue())
	if err != nil {
		return nil, fmt.Errorf("'%s' could not be parsed", timeStampPath)
	}
	dr := &JsonData{dataMap: rO, navIndex: createNavIndex(rO), dataMapUpdated: dataMapUpdated}
	return dr, nil
}

func (p *JsonData) GetNavIndex(id string) []string {
	ni := *p.navIndex
	return ni[id]
}

func (p *JsonData) GetDataMap() *parser.JsonObject {
	return p.dataMap
}

func (p *JsonData) GetNavIndexAsString() string {
	var sb strings.Builder
	var root string
	list := make([]string, 0)
	for n, v := range *p.navIndex {
		if n == "" {
			root = vToStr(v)
		} else {
			list = append(list, fmt.Sprintf("\n%s = \n%s", n, vToStr(v)))
		}
	}
	sort.Strings(list)
	sb.WriteString(root)
	for _, n := range list {
		sb.WriteString(n)
	}
	return sb.String()
}

func vToStr(v []string) string {
	var sb strings.Builder
	for _, v := range v {
		sb.WriteString(fmt.Sprintf("    %s,\n", v))
	}
	return strings.Trim(sb.String(), "\n")
}

/*
	We know we can cast the groups node. We checked it in NewJsonData

	Returns the node containing all of the users (groups).
*/
func (p *JsonData) GetUserRoot() *parser.JsonObject {
	return p.dataMap.GetNodeWithName(dataMapRootName).(*parser.JsonObject)
}

/*
	Returns the user id of the first and possibly only user.
	Used to select a user on first load
*/
func (p *JsonData) GetFirstUserName() string {
	return p.GetUserRoot().GetSortedKeys()[0]
}

/*
	Return the node of a given user
*/
func (p *JsonData) getUserNode(userName string) *parser.JsonObject {
	u := p.GetUserRoot().GetNodeWithName(userName)
	if u == nil {
		return nil
	}
	return u.(*parser.JsonObject)
}

func (p *JsonData) SetDateTime() {
	dtNode, err := parser.Find(p.dataMap, timeStampPath)
	if err == nil {
		dtNode.(*parser.JsonString).SetValue(time.Now().Format(dateTimeFormatStr))
	}
}

func (p *JsonData) GetTimeStampString() string {
	dtNode, err := parser.Find(p.dataMap, timeStampPath)
	if err == nil {
		return dtNode.String()
	}
	return "Undefined"
}

func (p *JsonData) ToJson() string {
	return p.dataMap.JsonValue()
}

func (p *JsonData) Search(addPath func(*parser.Path, string), needle string, matchCase bool) {
	groups := p.dataMap.GetNodeWithName(dataMapRootName).(*parser.JsonObject)
	for _, v := range groups.GetValues() {
		searchUsers(addPath, needle, v.GetName(), v.(*parser.JsonObject), matchCase)
	}
}

func (p *JsonData) CloneHint(dataPath *parser.Path, hintItemName string, cloneLeafNodeData bool) error {
	h, err := parser.Find(p.GetUserRoot(), dataPath)
	if err != nil {
		return fmt.Errorf("the clone hint '%s' cannot be found", dataPath)
	}
	if h.GetNodeType() != parser.NT_OBJECT {
		return fmt.Errorf("the clone hint item '%s' was not an object node", dataPath)
	}
	parent, _ := parser.FindParentNode(p.dataMap, h)
	if parent.(parser.NodeC).GetNodeWithName(hintItemName) != nil {
		return fmt.Errorf("the cloned item '%s' name already exists", dataPath)
	}
	cl := parser.Clone(h, hintItemName, cloneLeafNodeData)
	parent.(parser.NodeC).Add(cl)
	p.navIndex = createNavIndex(p.dataMap)
	p.dataMapUpdated(fmt.Sprintf("Cloned Item '%s' added", hintItemName), dataPath.PathFirst(2).StringAppend(hintItemName), nil)
	return nil
}

func (p *JsonData) AddSubItem(dataPath *parser.Path, hintItemName string, itemName string) error {
	h, err := parser.Find(p.GetUserRoot(), dataPath)
	if err != nil {
		return fmt.Errorf("the %s '%s' cannot be found", itemName, dataPath)
	}
	if h.GetNodeType() != parser.NT_OBJECT {
		return fmt.Errorf("the %s item '%s' was not an object node", itemName, dataPath)
	}
	hO := h.(*parser.JsonObject)
	addStringIfDoesNotExist(hO, hintItemName)
	p.navIndex = createNavIndex(p.dataMap)
	p.dataMapUpdated("Add Sub Item", dataPath.StringAppend(hintItemName), nil)
	return nil
}

func (p *JsonData) AddAsset(userPath *parser.Path, assetName string) error {
	u := p.getUserNode(userPath.StringFirst()) // User id is first path element
	if u == nil {
		return fmt.Errorf("the user '%s' cannot be found", userPath)
	}
	addAssetToUser(u, assetName)
	p.navIndex = createNavIndex(p.dataMap)
	p.dataMapUpdated("Add Asset", userPath.StringAppend(IdAssets).StringAppend(assetName), nil)
	return nil
}

func (p *JsonData) AddTransaction(dataPath *parser.Path, date time.Time, ref string, amount float64, txType TransactionTypeEnum) error {
	data, err := parser.Find(p.GetUserRoot(), dataPath)
	if err != nil {
		data, err = parser.Find(p.GetUserRoot(), dataPath.PathParent())
		if err != nil {
			return fmt.Errorf("the transaction node for '%s' cannot be found", dataPath)
		}
		addDefaultAccountItemsToAsset(data.(*parser.JsonObject))
		data, err = parser.Find(p.GetUserRoot(), dataPath)
		if err != nil {
			return fmt.Errorf("the transaction node for '%s' cannot be found", dataPath)
		}
	}
	addTransactionToAsset(data.(*parser.JsonList), newTranactionData(date, amount, ref, txType, data))
	p.navIndex = createNavIndex(p.dataMap)
	p.dataMapUpdated("Add Transaction", dataPath.PathParent(), nil)
	return nil
}

func (p *JsonData) AddHint(userUid *parser.Path, hintName string) error {
	u := p.getUserNode(userUid.StringFirst()) // User id is first path element
	if u == nil {
		return fmt.Errorf("the user '%s' cannot be found", userUid)
	}
	addHintToUser(u, hintName)
	p.navIndex = createNavIndex(p.dataMap)
	p.dataMapUpdated("Add Hint", userUid.StringAppend(IdHints).StringAppend(hintName), nil)
	return nil
}

func (p *JsonData) AddNoteItem(userUid *parser.Path, itemName string) error {
	u := p.getUserNode(userUid.StringFirst()) // User id is first path element
	if u == nil {
		return fmt.Errorf("the user '%s' cannot be found", userUid)
	}
	addNoteToUser(u, itemName, "")
	p.navIndex = createNavIndex(p.dataMap)
	p.dataMapUpdated("Add Note Item", userUid.StringAppend(IdNotes), nil)
	return nil
}

func (p *JsonData) AddUser(userName string) error {
	userPath := parser.NewBarPath(userName)
	u := p.getUserNode(userPath.StringFirst()) // User id is first path element
	if u != nil {
		return fmt.Errorf("the user '%s' already exists", userName)
	}
	userO := parser.NewJsonObject(userName)
	addNoteToUser(userO, "note", "text")
	addHintToUser(userO, "App1")
	p.GetUserRoot().Add(userO)
	p.navIndex = createNavIndex(p.dataMap)
	p.dataMapUpdated("Add User", userPath, nil)
	return nil
}

func (p *JsonData) Rename(dataPath *parser.Path, newName string) error {
	n, err := p.GetNodeForUserPath(dataPath)
	if err != nil {
		return fmt.Errorf("the item to rename '%s' was not found in the data", dataPath)
	}
	parent, ok := parser.FindParentNode(p.dataMap, n)
	if !ok {
		return fmt.Errorf("the item to rename '%s' does not have a valid parent", dataPath)
	}
	err = parser.Rename(p.dataMap, n, newName)
	if err != nil {
		return fmt.Errorf("rename '%s' failed. Error: '%s'", dataPath, err.Error())
	}
	p.navIndex = createNavIndex(p.dataMap)
	if parent.GetName() == dataMapRootName { // If the parent is groups then the user was renamed
		p.dataMapUpdated(fmt.Sprintf("Renamed User '%s'", n.GetName()), parser.NewBarPath(newName), nil)
	} else {
		p.dataMapUpdated(fmt.Sprintf("Renamed Item '%s'", n.GetName()), dataPath.PathParent().StringAppend(newName), nil)
	}
	return nil
}

func (p *JsonData) Remove(dataPath *parser.Path, min int) error {
	n, err := p.GetNodeForUserPath(dataPath)
	if err != nil {
		return fmt.Errorf("the item to remove '%s' was not found in the data", dataPath)
	}
	parent, ok := parser.FindParentNode(p.dataMap, n)
	if !ok {
		return fmt.Errorf("the item to remove '%s' does not have a valid parent", dataPath)
	}
	if parent.GetNodeType() != parser.NT_OBJECT {
		return fmt.Errorf("the item to remove '%s' does not have an object parent", dataPath)
	}
	parentObj := parent.(*parser.JsonObject)
	count := parentObj.Len()
	if count <= min {
		return fmt.Errorf("there must be at least %d element(s) remaining in this item", min)
	}
	parser.Remove(p.dataMap, n)
	if min < 0 && parentObj.Len() == 0 {
		parser.Remove(p.dataMap, parentObj)
	}
	p.navIndex = createNavIndex(p.dataMap)
	if parent.GetName() == dataMapRootName { // If the parent is groups then the user was renamed
		p.dataMapUpdated(fmt.Sprintf("Removed User '%s'", n.GetName()), parser.NewBarPath(""), nil)
	} else {
		p.dataMapUpdated(fmt.Sprintf("Removed Item '%s'", n.GetName()), dataPath.PathParent(), nil)
	}
	return nil
}

func (p *JsonData) IsStringNode(n parser.NodeI) bool {
	return n.GetNodeType() == parser.NT_STRING
}

func (p *JsonData) GetNodeForUserPath(dataPath *parser.Path) (parser.NodeI, error) {
	return GetNodeForUserPath(p.GetDataMap(), dataPath)
}

func GetNodeForUserPath(root parser.NodeI, dataPath *parser.Path) (parser.NodeI, error) {
	path := dataMapRootPath.PathAppend(dataPath)
	nodes, err := parser.Find(root, path)
	if err != nil {
		return nil, err
	}
	if nodes == nil {
		return nil, fmt.Errorf("nil returned from parser.Find(%s)", path)
	}
	return nodes, nil
}

/**
Validate the names of entities. These result in JSON entity names so require
some restrictions.
*/
func ProcessEntityName(entry string, nt NodeAnnotationEnum) (string, error) {
	if len(entry) == 0 {
		return "", fmt.Errorf("input is undefined")
	}
	if len(entry) < 2 {
		return "", fmt.Errorf("input '%s' is too short. Must be longer that 1 char", entry)
	}
	lcEntry := strings.ToLower(entry)
	for _, c := range lcEntry {
		if c < ' ' {
			return "", fmt.Errorf("input cannot not contain control characters")
		}
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (strings.ContainsRune(allowedCharsInName, c)) {
			continue
		}
		return "", fmt.Errorf("input must not contain character '%c'. Only '0..9', 'a..z', 'A..Z' and '%s' chars are allowed", c, allowedCharsInName)
	}
	return GetNodeAnnotationNameWithPrefix(nt, entry), nil
}

func searchUsers(addPath func(*parser.Path, string), needle, user string, m *parser.JsonObject, matchCase bool) {
	for _, v := range m.GetValues() {
		if v.GetName() == IdHints {
			for _, v1 := range v.(*parser.JsonObject).GetValues() {
				searchLeafNodes(addPath, true, needle, user, v1.GetName(), v1.(*parser.JsonObject), matchCase)
			}
		} else {
			searchLeafNodes(addPath, false, needle, user, v.GetName(), v.(*parser.JsonObject), matchCase)
		}
	}
}

func SearchNodesWithName(name string, m *parser.JsonObject, f func(node, parent parser.NodeI)) {
	parser.WalkNodeTree(m, nil, func(node, parent, target parser.NodeI) bool {
		if (node != nil) && (parent != nil) {
			if node.GetName() == name {
				f(node, parent)
			}
		}
		return false
	})
}

func searchLeafNodes(addPath func(*parser.Path, string), isHint bool, needle, user, name string, m parser.NodeC, matchCase bool) {
	tag1 := PATH_SEP
	if isHint {
		tag1 = PATH_SEP + IdHints + PATH_SEP
	}
	if containsWithCase(name, needle, matchCase) {
		addPath(parser.NewBarPath(user+tag1+name), searchDeriveText(user, isHint, name, "LHS Tree", ""))
	}
	for _, s := range m.GetValues() {
		if containsWithCase(s.GetName(), needle, matchCase) {
			addPath(parser.NewBarPath(user+tag1+name), searchDeriveText(user, isHint, name, "Field Name", s.GetName()))
		}
		if s.GetNodeType() == parser.NT_STRING {
			if containsWithCase(s.(*parser.JsonString).GetValue(), needle, matchCase) {
				addPath(parser.NewBarPath(user+tag1+name), searchDeriveText(user, isHint, name, "In Text", s.GetName()))
			}
		} else {
			if s.IsContainer() {
				searchLeafNodes(addPath, isHint, needle, user, s.GetName(), s.(parser.NodeC), matchCase)
			}
		}
	}
}

func searchDeriveText(user string, isHint bool, name, desc, key string) string {
	if key != "" {
		key = "'" + key + "'"
	}
	if isHint {
		return user + " [Hint] " + name + ":  " + desc + ": " + key
	}
	if name == "notes" {
		return user + " [Notes] :  " + desc + ": " + key
	}
	return user + " " + name + ":  " + desc + ": " + key
}

func addDefaultHintItemsToHint(hint *parser.JsonObject) {
	for _, n := range defaultHintNames {
		addStringIfDoesNotExist(hint, n)
	}
}

func addDefaultAccountItemsToAsset(account *parser.JsonObject) {
	for _, n := range defaultAssetNames {
		addStringIfDoesNotExist(account, n)
	}
	tx := account.GetNodeWithName(IdTransactions)
	if tx == nil {
		txl := parser.NewJsonList(IdTransactions)
		addTransactionToAsset(txl, newTranactionData(time.Now(), 0.0, "Opening Balance", TX_TYPE_IV, txl))
		account.Add(txl)
	}
}

func addTransactionToAsset(transactions *parser.JsonList, tx *TranactionData) {
	txo := parser.NewJsonObject("")
	txo.Add(parser.NewJsonString(IdTxDate, tx.DateTime()))
	txo.Add(parser.NewJsonString(IdTxRef, tx.ref))
	txo.Add(parser.NewJsonNumber(IdTxVal, tx.value))
	txo.Add(parser.NewJsonString(IdTxType, string(tx.txType)))
	transactions.Add(txo)
}

func addStringIfDoesNotExist(obj *parser.JsonObject, name string) {
	node := obj.GetNodeWithName(name)
	if node == nil {
		node = parser.NewJsonString(name, "")
		obj.Add(node)
	}
}

func addAssetToUser(userO *parser.JsonObject, assetName string) {
	assets := userO.GetNodeWithName(IdAssets)
	if assets == nil {
		assets = parser.NewJsonObject(IdAssets)
		userO.Add(assets)
	}
	acc0 := assets.(*parser.JsonObject)
	acc := acc0.GetNodeWithName(assetName)
	if acc == nil {
		acc = parser.NewJsonObject(assetName)
		acc0.Add(acc)
	}
	acc0 = acc.(*parser.JsonObject)
	addDefaultAccountItemsToAsset(acc0)
}

func addHintToUser(userO *parser.JsonObject, hintName string) {
	hints := userO.GetNodeWithName(IdHints)
	if hints == nil {
		hints = parser.NewJsonObject(IdHints)
		userO.Add(hints)
	}
	hintsO := hints.(*parser.JsonObject)
	hint := hintsO.GetNodeWithName(hintName)
	if hint == nil {
		hint = parser.NewJsonObject(hintName)
		hintsO.Add(hint)
	}
	hintO := hint.(*parser.JsonObject)
	addDefaultHintItemsToHint(hintO)
}

func addNoteToUser(userO *parser.JsonObject, noteName, noteText string) {
	notes := userO.GetNodeWithName(IdNotes)
	if notes == nil {
		notes = parser.NewJsonObject(IdNotes)
		userO.Add(notes)
	}
	notesO := notes.(*parser.JsonObject)
	note := notesO.GetNodeWithName(noteName)
	if note == nil {
		note = parser.NewJsonString(noteName, noteText)
		notesO.Add(note)
	}
}

func containsWithCase(haystack, needle string, matchCase bool) bool {
	if matchCase {
		return strings.Contains(haystack, needle)
	} else {
		h := strings.ToLower(haystack)
		n := strings.ToLower(needle)
		return strings.Contains(h, n)
	}
}

func createNavIndex(m parser.NodeI) *map[string][]string {
	var uids = make(map[string][]string)
	createNavIndexDetail("", &uids, m)

	return &uids
}

func createNavIndexDetail(id string, uids *map[string][]string, nodeI parser.NodeI) {
	if nodeI.GetNodeType() == parser.NT_LIST {
		panic("createNavIndexDetail2: cannot process lists")
	}
	g, err := parser.Find(nodeI, dataMapRootPath)
	if err != nil {
		panic("createNavIndexDetail2: cannot find groups")
	}
	userObj := g.(*parser.JsonObject)
	(*uids)[""] = userObj.GetSortedKeys()
	users := userObj.GetValuesSorted()
	for _, user := range users {
		if user.GetNodeType() == parser.NT_OBJECT {
			userO := user.(*parser.JsonObject)
			id := userO.GetName()
			l, ll := keysToList(id, userO)
			(*uids)[id] = ll
			for ii2, id2 := range l {
				m2 := userO.GetNodeWithName(id2).(*parser.JsonObject)
				for _, v2 := range m2.GetValuesSorted() {
					if v2.GetNodeType() == parser.NT_OBJECT {
						l2, ll2 := keysToList(ll[ii2], m2)
						(*uids)[ll[ii2]] = ll2
						for ii3, id3 := range l2 {
							m3 := m2.GetNodeWithName(id3)
							if m3.GetNodeType() == parser.NT_OBJECT {
								m3O := m3.(*parser.JsonObject)
								for _, v3 := range m3O.GetValuesSorted() {
									if v3.GetNodeType() == parser.NT_OBJECT {
										an, _ := GetNodeAnnotationTypeAndName(v3.GetName())
										if an == NOTE_TYPE_SL {
											l3 := make([]string, 1)
											l3[0] = v3.GetName()
											(*uids)[ll2[ii3]] = l3
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func keysToList(id string, m *parser.JsonObject) ([]string, []string) {
	l := make([]string, 0)
	ll := make([]string, 0)
	for _, k := range m.GetSortedKeys() {
		l = append(l, k)
		if id == "" {
			ll = append(ll, k)
		} else {
			ll = append(ll, fmt.Sprintf("%s%s%s", id, PATH_SEP, k))
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

func GetLastId(uid string) string {
	if uid == "" {
		return uid
	}
	p := strings.LastIndexByte(uid, PATH_SEP_CHAR)
	switch p {
	case -1:
		return uid
	case 0:
		return uid[1:]
	default:
		return uid[p+1:]
	}
}

func GetUidPathFromDataPath(dataPath *parser.Path) *parser.Path {
	if dataPath.Len() == 0 {
		return dataPath
	}
	if dataPath.StringFirst() == dataMapRootName {
		dataPath = dataPath.PathLast(dataPath.Len() - 1)
	}
	return dataPath
}

func GetParentId(uid string) string {
	if uid == "" {
		return uid
	}
	p := strings.LastIndexByte(uid, PATH_SEP_CHAR)
	switch p {
	case -1:
		return uid
	case 0:
		return ""
	default:
		return uid[0:p]
	}
}

func GetPathElementAt(path string, index int) string {
	elements := strings.Split(path, PATH_SEP)
	l := len(elements)
	if l == 0 || index < 0 || index >= l {
		return ""
	}
	if index < l {
		return elements[index]
	}
	return elements[l-1]
}

func GetUserFromUid(uid string) string {
	return GetFirstUidElements(uid, 1)
}

func GetHintFromUid(uid string) string {
	return GetPathElementAt(uid, 2)
}

func GetFirstUidElements(uid string, count int) string {
	if count <= 0 {
		return ""
	}
	var sb strings.Builder
	dotCount := 0
	for _, c := range uid {
		if c == PATH_SEP_CHAR {
			dotCount++
		}
		if dotCount == count {
			return sb.String()
		}
		sb.WriteByte(byte(c))
	}
	return sb.String()
}

func (p *JsonData) GetRootUidOrCurrentUid(currentUid string) string {
	ni := *p.navIndex
	if currentUid != "" {
		for i := 4; i > 0; i-- {
			x := GetFirstUidElements(currentUid, i)
			_, ok := ni[x]
			if ok {
				return currentUid
			}
		}
	}
	l := make([]string, 0)
	for k := range ni {
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

func IndexOfAnnotation(annotation string) int {
	for i, v := range nodeAnnotationPrefix {
		if v == annotation {
			return i
		}
	}
	return 0
}

func GetNodeAnnotationTypeAndName(combinedName string) (NodeAnnotationEnum, string) {
	pos := strings.IndexRune(combinedName, '!')
	if pos < 0 {
		return NOTE_TYPE_SL, combinedName
	}
	aStr := combinedName[pos:]
	indx := IndexOfAnnotation(aStr)
	if indx == 0 {
		return NOTE_TYPE_SL, combinedName
	}
	return NodeAnnotationEnums[indx], combinedName[:pos]
}

func GetNodeAnnotationNameWithPrefix(nae NodeAnnotationEnum, name string) string {
	return name + nodeAnnotationPrefix[nae]
}

func GetNodeAnnotationPrefixName(nae NodeAnnotationEnum) string {
	return nodeAnnotationPrefix[nae]
}

func CreateEmptyJsonData() []byte {
	root := parser.NewJsonObject("")
	g := parser.NewJsonObject(dataMapRootName)
	u := parser.NewJsonObject("tempUser")
	n := parser.NewJsonObject(IdNotes)
	h := parser.NewJsonObject(IdHints)
	n1 := parser.NewJsonString("note", "newNote")
	app := parser.NewJsonObject("application")
	name := parser.NewJsonString("name", "userName")
	app.Add(name)
	n.Add(n1)
	u.Add(n)
	u.Add(h)
	g.Add(u)
	h.Add(app)
	root.Add(g)
	root.Add(parser.NewJsonString(timeStampName, time.Now().Format(dateTimeFormatStr)))
	return []byte(root.JsonValueIndented(4))
}
