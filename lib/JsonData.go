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
	// HI Remove Note
	IdHints            = "pwHints"
	DataMapRootName    = "groups"
	timeStampName      = "timeStamp"
	tabdata            = "                                     "
	allowedCharsInName = " *@#$%^&*()_+=?"
	dateTimeFormatStr  = "Mon Jan 2 15:04:05 MST 2006"
	PATH_SEP           = "|"
	PATH_SEP_CHAR      = '|'

	NODE_TYPE_SL NodeAnnotationEnum = 0 // Single Line: These are indexes. Found issues when using iota!
	NODE_TYPE_ML NodeAnnotationEnum = 1 // Multi Line
	NODE_TYPE_RT NodeAnnotationEnum = 2 // Rich Text
	NODE_TYPE_PO NodeAnnotationEnum = 3 // POsitinal
	NODE_TYPE_IM NodeAnnotationEnum = 4 // IMage
)

var (
	nodeAnnotationPrefix      = []string{"", "!ml", "!rt", "!po", "!im"}
	NodeAnnotationPrefixNames = []string{"Single Line", "Multi Line", "Rich Text", "Positional", "Image"}
	NodeAnnotationEnums       = []NodeAnnotationEnum{NODE_TYPE_SL, NODE_TYPE_ML, NODE_TYPE_RT, NODE_TYPE_PO, NODE_TYPE_IM}
	NodeAnnotationsSingleLine = []bool{true, false, false, true, true}
	defaultHintNames          = []string{"notes", "post", "pre", "userId"}
	defaultAssetNames         = []string{"Account Num.", "Sort Code"}
	timeStampPath             = parser.NewBarPath(timeStampName)
	dataMapRootPath           = parser.NewBarPath(DataMapRootName)
	nameMap                   = make(map[string]string)
)

type JsonData struct {
	dataMap        *parser.JsonObject
	navIndex       *map[string][]string
	dataMapUpdated func(string, *parser.Path, error)
}

func InitNameMap(m map[string]string) {
	nameMap = m
}

func GetNameFromNameMap(key string, fb string) string {
	val, found := nameMap[key]
	if found {
		return val
	}
	if fb == "" {
		return key
	}
	return fb
}

//
// Return the path after the DataMapRootName.
//
func GetPathAfterDataRoot(dataPath *parser.Path) *parser.Path {
	for i := 0; i < dataPath.Len(); i++ {
		if dataPath.StringAt(i) == DataMapRootName {
			return dataPath.PathLast(dataPath.Len() - (i + 1))
		}
	}
	return dataPath
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

	u := rO.GetNodeWithName(DataMapRootName)
	if u == nil {
		return nil, fmt.Errorf("root '%s' element is missing", DataMapRootName)
	}
	if u.GetNodeType() != parser.NT_OBJECT {
		return nil, fmt.Errorf("root '%s' element is NOT a JsonObject", DataMapRootName)
	}
	ts, err := parser.Find(rO, timeStampPath)
	if err != nil {
		return nil, fmt.Errorf("'%s' does not exist in data root", timeStampPath)
	}
	_, err = parseTime((ts.(*parser.JsonString)).GetValue())
	if err != nil {
		return nil, fmt.Errorf("'%s' could not be parsed", timeStampPath)
	}

	nl := make([]parser.NodeI, 0)
	for _, n := range u.(parser.NodeC).GetValues() {
		parser.WalkNodeTreeForTrail(n.(parser.NodeC), func(t *parser.Trail, i int) bool {
			if t.Len() == 1 {
				nn := t.GetNodeAt(0).(parser.NodeC)
				if nn.Len() > 0 && nn.GetValues()[0].GetNodeType() == parser.NT_STRING {
					nl = append(nl, nn)
				}
			}
			return false
		})
	}
	for _, xx := range nl {
		fmt.Printf("REM: %s\n", xx.String())
		err := parser.Remove(u, xx)
		if err != nil {
			fmt.Printf("ERR: %s\n", err.Error())
		}
	}

	dr := &JsonData{dataMap: rO, navIndex: createNavIndex(rO), dataMapUpdated: dataMapUpdated}
	return dr, nil
}

func (p *JsonData) GetNavIndex(id string) []string {
	ni := *p.navIndex
	return ni[id]
}

func (p *JsonData) GetDataRoot() *parser.JsonObject {
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
	u := p.dataMap.GetNodeWithName(DataMapRootName)
	if u == nil {
		return nil
	}
	return u.(*parser.JsonObject)
}

/*
	Returns the user id of the first and possibly only user.
	Used to select a user on first load
*/
func (p *JsonData) GetFirstUserName() string {
	u := p.GetUserRoot()
	if u == nil {
		return ""
	}
	l := u.GetSortedKeys()
	if len(l) != 0 {
		return l[0]
	}
	return ""
}

/*
	Return the node of a given user
*/
func (p *JsonData) getUserNode(userName string) *parser.JsonObject {
	u := p.GetUserRoot()
	if u == nil {
		return nil
	}
	n := u.GetNodeWithName(userName)
	if n == nil {
		return nil
	}
	if n.GetNodeType() != parser.NT_OBJECT {
		return nil
	}
	return n.(*parser.JsonObject)
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

func (p *JsonData) Search(addTrailFunc func(*parser.Trail), needle string, ignoreCase bool) {
	dataRoot := p.dataMap.GetNodeWithName(DataMapRootName).(*parser.JsonObject)
	searchGroups(addTrailFunc, needle, dataRoot, ignoreCase)
}

func (p *JsonData) AddTransaction(transactionPath *parser.Path, date time.Time, ref string, amount float64, txType TransactionTypeEnum) error {
	userRoot := p.GetUserRoot()
	if userRoot == nil {
		return fmt.Errorf("the user root node cannot be found")
	}
	txNode, err := parser.Find(userRoot, transactionPath)
	if err != nil {
		return fmt.Errorf("the transaction node for '%s' cannot be found", transactionPath)
	}
	if txNode.GetNodeType() != parser.NT_LIST {
		return fmt.Errorf("the transaction node for '%s' is not a List node", transactionPath)
	}
	addTransactionToAsset(txNode.(*parser.JsonList), newTranactionData(date, amount, ref, txType, txNode))
	p.navIndex = createNavIndex(p.dataMap)
	p.dataMapUpdated("Add Transaction", transactionPath.PathParent(), nil)
	return nil
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

func (p *JsonData) AddSubItem(dataPath *parser.Path, subItemName string, itemDisplayName string) error {
	h, err := parser.Find(p.GetUserRoot(), dataPath)
	if err != nil {
		return fmt.Errorf("the item '%s' cannot be found", dataPath)
	}
	if h.GetNodeType() != parser.NT_OBJECT {
		return fmt.Errorf("the item '%s' is not a Json Object", dataPath)
	}
	hO := h.(*parser.JsonObject)
	ok := addStringIfDoesNotExist(hO, subItemName)
	if ok {
		p.navIndex = createNavIndex(p.dataMap)
		p.dataMapUpdated("Add Sub Item", dataPath.StringAppend(subItemName), nil)
		return nil
	}
	return fmt.Errorf("the item '%s' already contains '%s'", dataPath, subItemName)
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

func (p *JsonData) AddUser(userName string) error {
	u := p.getUserNode(userName) // User id is first path element
	if u != nil {
		return fmt.Errorf("the user '%s' already exists", userName)
	}
	userO := parser.NewJsonObject(userName)
	addHintToUser(userO, "App1")
	p.GetUserRoot().Add(userO)
	p.navIndex = createNavIndex(p.dataMap)
	p.dataMapUpdated("Add User", parser.NewDotPath(userName), nil)
	return nil
}

func (p *JsonData) Rename(dataPath *parser.Path, newName string) error {
	n, err := p.FindNodeForUserDataPath(dataPath)
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
	if parent.GetName() == DataMapRootName { // If the parent is groups then the user was renamed
		p.dataMapUpdated(fmt.Sprintf("Renamed User '%s'", n.GetName()), parser.NewBarPath(newName), nil)
	} else {
		p.dataMapUpdated(fmt.Sprintf("Renamed Item '%s'", n.GetName()), dataPath.PathParent().StringAppend(newName), nil)
	}
	return nil
}

func (p *JsonData) Remove(dataPath *parser.Path, min int) error {
	n, err := p.FindNodeForUserDataPath(dataPath)
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
	if parent.GetName() == DataMapRootName { // If the parent is groups then the user was renamed
		p.dataMapUpdated(fmt.Sprintf("Removed User '%s'", n.GetName()), parser.NewBarPath(""), nil)
	} else {
		p.dataMapUpdated(fmt.Sprintf("Removed Item '%s'", n.GetName()), dataPath.PathParent(), nil)
	}
	return nil
}

func (p *JsonData) FindNodeForUserDataPath(dataPath *parser.Path) (parser.NodeI, error) {
	return FindNodeForUserDataPath(p.GetDataRoot(), dataPath)
}

//
// User data path is a path where the first node is the user. As this is NOT the root
//  of the json data we append the patg to the data root (usually 'groups') to give
//  a full path and the we Find the node using the path
//
func FindNodeForUserDataPath(root parser.NodeI, dataPathForUser *parser.Path) (parser.NodeI, error) {
	path := dataMapRootPath.PathAppend(dataPathForUser)
	nodes, err := parser.Find(root, path)
	if err != nil {
		return nil, err
	}
	if nodes == nil {
		return nil, fmt.Errorf("FindNodeForUserDataPath: nil returned from parser.Find(%s)", path)
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

func searchGroups(addTrailFunc func(*parser.Trail), needle string, m *parser.JsonObject, ignoreCase bool) {
	if ignoreCase {
		needle = strings.ToLower(needle)
	}
	parser.WalkNodeTreeForTrail(m, func(t *parser.Trail, i int) bool {
		last := t.GetLast()
		if last != nil {
			name := last.GetName()
			if name != "" {
				if ignoreCase {
					name = strings.ToLower(name)
				}
				if strings.Contains(name, needle) {
					addTrailFunc(t)
				}
			}
			if !last.IsContainer() {
				value := last.String()
				if value != "" {
					if ignoreCase {
						value = strings.ToLower(value)
					}
					if strings.Contains(value, needle) {
						addTrailFunc(t)
					}
				}
			}
		}
		return false
	})
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

func addStringIfDoesNotExist(obj *parser.JsonObject, name string) bool {
	node := obj.GetNodeWithName(name)
	if node == nil {
		node = parser.NewJsonString(name, "")
		obj.Add(node)
		return true
	}
	return false
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
										if an == NODE_TYPE_SL {
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
		return NODE_TYPE_SL, combinedName
	}
	aStr := combinedName[pos:]
	indx := IndexOfAnnotation(aStr)
	if indx == 0 {
		return NODE_TYPE_SL, combinedName
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
	g := parser.NewJsonObject(DataMapRootName)
	u := parser.NewJsonObject("tempUser")
	h := parser.NewJsonObject(IdHints)
	app := parser.NewJsonObject("application")
	name := parser.NewJsonString("name", "userName")
	app.Add(name)
	u.Add(h)
	g.Add(u)
	h.Add(app)
	root.Add(g)
	root.Add(parser.NewJsonString(timeStampName, time.Now().Format(dateTimeFormatStr)))
	return []byte(root.JsonValueIndented(4))
}
