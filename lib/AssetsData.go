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
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stuartdd2/JsonParser4go/parser"
)

type TransactionTypeEnum string

const (
	TIME_FORMAT_TXN = "2006-01-02 15:04:05"
	TIME_FORMAT_CSV = "02/01/2006"
	IdTransactions  = "transactions"
	IdAssets        = "assets"
	IdTxDate        = "date"
	IdTxRef         = "ref"
	IdTxVal         = "val"
	IdTxType        = "type"

	TX_TYPE_ERR TransactionTypeEnum = "err"
	TX_TYPE_IV  TransactionTypeEnum = "iv"
	TX_TYPE_DEB TransactionTypeEnum = "db"
	TX_TYPE_CRE TransactionTypeEnum = "cr"
)

var (
	TX_TYPE_LIST_LABLES     = []string{"In (Credit)", "Out (Debit)"}
	TX_TYPE_LIST_OPTIONS    = []string{string(TX_TYPE_CRE), string(TX_TYPE_DEB)}
	IMPORT_CSV_MAP_LIST_DEF = []string{"date", "type", "", "", "ref", "val", "val", ""}
)

var cachedUserAssets *UserAssetCache

//
// UserAssetCache map[string]*UserAsset
//		[username+assetname] *UserAsset
//			user --> json user node
//			asset --> json asset node
//          	'groups'.username.'assets'.assetname
//			data --> []*AccountData
//					LatestTransaction() *TranactionData
//				AccountName --> assetname
//				Path --> 'groups'.username.'assets'.assetname
//				Transactions --> []*TranactionData
//					dateTime 	DateTime()
//					value		Value() float64
//								AbsValue() float64
//								Val() string - formatted value
//					ref			Ref() string
//					txType		TxType() TransactionTypeEnum
//								Key() string - date + ref
//								LineValue() float64
//								LineVal() string - formatted value
//								HasError() bool
//

type UserAssetCache struct {
	UserAssets map[string]*UserAsset
}

type UserAsset struct {
	user  parser.NodeC   // This is the asset parent node (The user)
	asset parser.NodeC   // The asset node. This contains the all assets for a user
	data  []*AccountData // List of accounts in assets (accounts, organisations, etc)
}

type TranactionData struct {
	dateTime  string
	value     float64
	ref       string
	txType    TransactionTypeEnum
	err       error
	lineValue float64
}

type AccountData struct {
	Path         parser.Path
	AccountName  string            // Like Lloyds Bank Current Account
	InitialValue float64           // initial value.
	ClosingValue float64           // initial value -+ all transactions
	Transactions []*TranactionData // Each transaction
}

//
// Create/Update the cachedUserAssets (singleton) from the root Json node.
//  All assets for All users
//
func InitUserAssetsCache(root *parser.JsonObject) {
	cache := &UserAssetCache{UserAssets: make(map[string]*UserAsset)}
	SearchNodesWithName(IdAssets, root, func(node, parent parser.NodeI) {
		if root.IsContainer() {
			if node.IsContainer() && parent.IsContainer() {
				cache.addAsset(newUserAsset(parent.(parser.NodeC), node.(parser.NodeC)))
			}
		}
	})
	cachedUserAssets = cache
}

func ImportCsvData(txNode parser.NodeC, fileName string, skipFirstLine bool, dtFormat string, mapList []string) error {
	if cachedUserAssets == nil {
		return fmt.Errorf("cannot import no assets have been defined")
	}
	return fmt.Errorf("cannot import no assets have been defined")
}

//
// Add a UserAsset to the cachedUserAssets (singleton) using key of username + assetname
//
func (t *UserAssetCache) addAsset(asset *UserAsset) {
	t.UserAssets[asset.keyForUserAsset()] = asset
}

//
// Return all accounts (AccountData) for a given user from the cachedUserAssets
//
func FindAllUserAccounts(user string) ([]*AccountData, error) {
	if cachedUserAssets == nil {
		return nil, fmt.Errorf("no assets or accounts have been defined")
	}
	key := fmt.Sprintf("%s|%s", user, IdAssets)
	ua, ok := cachedUserAssets.UserAssets[key]
	if ok {
		return ua.data, nil
	}
	return nil, fmt.Errorf("assets not found for user '%s'", user)
}

//
// Return an account (AccountData) for a given user name and account name from the cachedUserAssets
//
func FindUserAccount(user, account string) (*AccountData, error) {
	ua, err := FindAllUserAccounts(user)
	if err == nil {
		for _, acc := range ua {
			if acc.AccountName == account {
				return acc, nil
			}
		}
	}
	return nil, fmt.Errorf("account '%s' not found for user '%s'", account, user)
}

func StringUserAsset() string {
	if cachedUserAssets == nil {
		return "UserAssetCache:is nil"
	}
	var sb strings.Builder
	sb.WriteString("AssetCache:\n")
	for _, v := range cachedUserAssets.UserAssets {
		sb.WriteString(fmt.Sprintf("    %s,\n", v))
	}
	return strings.Trim(sb.String(), "\n")
}

func newUserAsset(userNode, assetsNode parser.NodeC) *UserAsset {
	ad := make([]*AccountData, 0)
	for _, accN := range assetsNode.GetValues() {
		if accN.IsContainer() {
			ad = append(ad, newAccountData(accN.(parser.NodeC), 0.0))
		}
	}
	return &UserAsset{user: userNode, asset: assetsNode, data: ad}
}

func (t *UserAsset) keyForUserAsset() string {
	return userAssetKey(t.user, t.asset)
}

func userAssetKey(userNode, assetsNode parser.NodeC) string {
	return fmt.Sprintf("%s|%s", userNode.GetName(), assetsNode.GetName())
}

// !tx node. List of all transactions.
// Sorted by datetime.
func newAccountData(accountNode parser.NodeC, initialValue float64) *AccountData {
	d := make([]*TranactionData, 0)
	v := initialValue
	for _, n := range accountNode.GetValues() {
		if n.GetName() == IdTransactions && n.IsContainer() {
			for _, ni := range n.(parser.NodeC).GetValues() {
				d = append(d, NewTranactionDataFromNode(ni))
			}
			sort.Slice(d, func(i, j int) bool {
				return getMillisForDateTime(d[i].dateTime) < getMillisForDateTime(d[j].dateTime)
			})
			for _, ni := range d {
				if ni.txType == TX_TYPE_DEB {
					ni.SetLineValue(v - ni.AbsValue())
				} else {
					ni.SetLineValue(v + ni.AbsValue())
				}
				v = ni.LineValue()
			}
		}
	}
	return &AccountData{AccountName: accountNode.GetName(), InitialValue: initialValue, ClosingValue: v, Transactions: d}
}

func (t *AccountData) String() string {
	return fmt.Sprintf("Name: %s Initial value:%9.2f Final value:%9.2f", t.AccountName, t.InitialValue, t.ClosingValue)
}

func (t *AccountData) LatestTransaction() *TranactionData {
	if len(t.Transactions) == 0 {
		return nil
	}
	var td int64 = 0
	ind := 0
	for i, t := range t.Transactions {
		if td < getMillisForDateTime(t.dateTime) {
			td = getMillisForDateTime(t.dateTime)
			ind = i
		}
	}
	return t.Transactions[ind]
}

//
// Give node (n) must be a 'transactions' container node
//	return a sub node that has a date and ref the san=me as datePlusRef
//
func GetTransactionNode(n parser.NodeC, datePlusRef string) (parser.NodeC, error) {
	if n.GetName() != IdTransactions {
		return nil, fmt.Errorf("GetTransaction failed. Node is not '%s'", IdTransactions)
	}
	for _, t := range n.GetValues() {
		if t.IsContainer() {
			tc := t.(parser.NodeC)
			tdt := tc.GetNodeWithName(IdTxDate)
			if tdt == nil {
				return nil, fmt.Errorf("failed GetTransaction. '%s' member not found", IdTxDate)
			}
			tref := tc.GetNodeWithName(IdTxRef)
			if tref == nil {
				return nil, fmt.Errorf("failed GetTransaction. '%s' member not found", tref)
			}
			s := fmt.Sprintf("%s %s", tdt, tref)
			if s == datePlusRef {
				return tc, nil
			}
		}
	}
	return nil, fmt.Errorf("failed GetTransaction. '%s' transaction not found", datePlusRef)
}

func newTranactionData(dateTime string, value float64, ref string, typ TransactionTypeEnum) *TranactionData {
	return &TranactionData{dateTime: dateTime, value: value, ref: ref, lineValue: 0.0, txType: typ}
}

func newTranactionDataError(err string, n parser.NodeI) *TranactionData {
	return &TranactionData{err: fmt.Errorf("%s '%s'", err, n.JsonValue()), lineValue: 0.0, txType: TX_TYPE_ERR}
}

func (t *TranactionData) DateTime() string {
	return t.dateTime
}

func (t *TranactionData) Key() string {
	return fmt.Sprintf("%s %s", t.dateTime, t.ref)
}

func (t *TranactionData) Ref() string {
	return t.ref
}

func (t *TranactionData) Val() string {
	return fmt.Sprintf("%9.2f", t.value)
}

func (t *TranactionData) TxType() TransactionTypeEnum {
	return t.txType
}

func (t *TranactionData) LineVal() string {
	return fmt.Sprintf("%9.2f", t.lineValue)
}

func (t *TranactionData) Value() float64 {
	return t.value
}
func (t *TranactionData) AbsValue() float64 {
	return math.Abs(t.value)
}

func (t *TranactionData) LineValue() float64 {
	return t.lineValue
}

func (t *TranactionData) SetLineValue(lineValue float64) {
	t.lineValue = lineValue
}

func (t *TranactionData) String() string {
	if t.HasError() {
		return t.err.Error()
	}
	return fmt.Sprintf("%s %s %s %s", t.DateTime(), t.Ref(), t.Val(), t.LineVal())
}

func (t *TranactionData) HasError() bool {
	return t.err != nil
}

func getMillisForDateTime(dt string) int64 {
	t, err := time.Parse(TIME_FORMAT_TXN, dt)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}

func CurrentDateString() string {
	return time.Now().Format(TIME_FORMAT_TXN)
}

func ParseDateString(dts string) (time.Time, error) {
	dt, err := time.Parse(TIME_FORMAT_TXN, dts)
	if err != nil {
		return time.Now(), err
	}
	return dt, nil
}

//
// Update a json transaction node from a key and value
// 		keys 'date', 'ref', 'val', 'type'
// If node is a number then value is a float as a string
// else node is a non empty string
//
//
func UpdateNodeFromTranactionData(txNode parser.NodeC, key, value string, txType TransactionTypeEnum) int {
	vn := txNode.GetNodeWithName(key)
	if vn == nil {
		return 0
	}
	switch vn.GetNodeType() {
	case parser.NT_NUMBER:
		v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err == nil {
			if v != vn.(*parser.JsonNumber).GetValue() {
				vn.(*parser.JsonNumber).SetValue(v)
				return 1
			}
		}
	case parser.NT_STRING:
		if value != vn.(*parser.JsonString).GetValue() {
			vn.(*parser.JsonString).SetValue(value)
			return 1
		}
	}
	return 0
}

//
// Create a transaction (TranactionData) from a json Node
//
func NewTranactionDataFromNode(n parser.NodeI) *TranactionData {
	if n.IsContainer() {
		tn := n.(parser.NodeC).GetNodeWithName(IdTxDate)
		if tn == nil {
			return newTranactionDataError(fmt.Sprintf("Invalid Transaction node has no '%s' member", IdTxDate), n)
		}
		_, err := time.Parse(TIME_FORMAT_TXN, tn.String())
		if err != nil {
			return newTranactionDataError("invalid 'date time' for transaction %s", n)
		}
		ty := n.(parser.NodeC).GetNodeWithName(IdTxType)
		tys := TX_TYPE_ERR
		if ty != nil && ty.GetNodeType() == parser.NT_STRING {
			tys = TransactionTypeEnum(ty.String())
		}
		vn := n.(parser.NodeC).GetNodeWithName("val")
		if vn == nil {
			return newTranactionDataError("invalid Transaction node has no 'val' member '%s'", n)
		}
		if vn.GetNodeType() != parser.NT_NUMBER {
			return newTranactionDataError("invalid Transaction node 'val' member is  not a number'%s'", n)
		}
		val := vn.(*parser.JsonNumber).GetValue()
		rn := n.(parser.NodeC).GetNodeWithName(IdTxRef)
		if rn == nil {
			return newTranactionDataError(fmt.Sprintf("Invalid Transaction node has no '%s' member", IdTxRef), n)
		}
		ref := rn.String()
		return newTranactionData(tn.String(), val, ref, tys)
	} else {
		return newTranactionDataError("invalid Transaction node has no members (date, val and ref) '%s'", n)
	}
}
