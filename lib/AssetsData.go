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
	"time"

	"github.com/stuartdd2/JsonParser4go/parser"
)

const (
	TIME_FORMAT_TXN_IN  = "2006-01-02T15:04:05"
	TIME_FORMAT_TXN_OUT = "2006-01-02 15:04:05"
	idTransactions      = "transactions"
)

type UserAssetCache struct {
	UserAssets map[string]*UserAsset
}

type UserAsset struct {
	user  parser.NodeC   // This is the asset parent node (The user)
	asset parser.NodeC   // The asset node. This contains the all assets for a user
	data  []*AccountData // List of accounts in assets (accounts, organisations, etc)
}

type TranactionData struct {
	dateTime  time.Time
	value     float64
	ref       string
	err       error
	lineValue float64
}

type AccountData struct {
	AccountName  string            // Like Lloyds Bank Current Account
	InitialValue float64           // initial value.
	ClosingValue float64           // initial value -+ all transactions
	Transactions []*TranactionData // Each transaction
}

func NewUserAssetCache() *UserAssetCache {
	return &UserAssetCache{UserAssets: make(map[string]*UserAsset)}
}

func (t *UserAssetCache) Add(asset *UserAsset) {
	t.UserAssets[asset.Key()] = asset
}

func (t *UserAssetCache) Find(user, assets, account string) *AccountData {
	key := fmt.Sprintf("%s|%s", user, assets)
	ua, ok := t.UserAssets[key]
	if ok {
		for _, acc := range ua.data {
			if acc.AccountName == account {
				return acc
			}
		}
	}
	return nil
}

func NewUserAsset(userNode, assetsNode parser.NodeC) *UserAsset {
	ad := make([]*AccountData, 0)
	for _, accN := range assetsNode.GetValues() {
		if accN.IsContainer() {
			ad = append(ad, NewAccountData(accN.GetName(), accN.(parser.NodeC), 0.0))
		}
	}
	return &UserAsset{user: userNode, asset: assetsNode, data: ad}
}

func (t *UserAsset) Key() string {
	return UserAssetKey(t.user, t.asset)
}

func UserAssetKey(userNode, assetsNode parser.NodeC) string {
	return fmt.Sprintf("%s|%s", userNode.GetName(), assetsNode.GetName())
}

// !tx node. List of all transactions.
// Sorted by datetime.
func NewAccountData(accountName string, accountNode parser.NodeC, initialValue float64) *AccountData {
	d := make([]*TranactionData, 0)
	v := initialValue
	for _, n := range accountNode.GetValues() {
		if n.GetName() == idTransactions && n.IsContainer() {
			for _, ni := range n.(parser.NodeC).GetValues() {
				d = append(d, NewTranactionDataFromNode(ni))
			}
			sort.Slice(d, func(i, j int) bool {
				return d[i].dateTime.UnixMilli() < d[j].dateTime.UnixMilli()
			})
			for _, ni := range d {
				ni.SetLineValue(v - ni.Value())
				v = ni.LineValue()
			}
		}
	}
	return &AccountData{AccountName: accountName, InitialValue: initialValue, ClosingValue: v, Transactions: d}
}

func (t *AccountData) String() string {
	return fmt.Sprintf("Name: %s Initial value:%9.2f Final value:%9.2f", t.AccountName, t.InitialValue, t.ClosingValue)
}

func NewTranactionData(dateTime time.Time, value float64, ref string) *TranactionData {
	return &TranactionData{dateTime: dateTime, value: value, ref: ref, lineValue: 0.0}
}

func NewTranactionDataError(err string, n parser.NodeI) *TranactionData {
	return &TranactionData{err: fmt.Errorf("%s '%s'", err, n.JsonValue()), lineValue: 0.0}
}

func (t *TranactionData) DateTime() string {
	return t.dateTime.Local().Format(TIME_FORMAT_TXN_OUT)
}

func (t *TranactionData) Ref() string {
	return t.ref
}

func (t *TranactionData) Val() string {
	return fmt.Sprintf("%9.2f", t.value)
}

func (t *TranactionData) GetDC() string {
	if t.value < 0 {
		return "C"
	}
	return "D"
}

func (t *TranactionData) LineVal() string {
	return fmt.Sprintf("%9.2f", t.lineValue)
}

func (t *TranactionData) Value() float64 {
	return t.value
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
	return fmt.Sprintf("%s %s %s %s %s", t.DateTime(), t.Ref(), t.GetDC(), t.Val(), t.LineVal())
}

func (t *TranactionData) HasError() bool {
	return t.err != nil
}

func NewTranactionDataFromNode(n parser.NodeI) *TranactionData {
	if n.IsContainer() {
		tn := n.(parser.NodeC).GetNodeWithName("date")
		if tn == nil {
			return NewTranactionDataError("Invalid Transaction node has no 'date' member", n)
		}
		tim, err := time.Parse(TIME_FORMAT_TXN_IN, tn.String())
		if err != nil {
			return NewTranactionDataError("invalid 'date' for transaction %s", n)
		}
		vn := n.(parser.NodeC).GetNodeWithName("val")
		if vn == nil {
			return NewTranactionDataError("invalid Transaction node has no 'val' member '%s'", n)
		}
		if vn.GetNodeType() != parser.NT_NUMBER {
			return NewTranactionDataError("invalid Transaction node 'val' member is  not a number'%s'", n)
		}
		val := vn.(*parser.JsonNumber).GetValue()
		rn := n.(parser.NodeC).GetNodeWithName("ref")
		if rn == nil {
			return NewTranactionDataError("invalid Transaction node has no 'ref' member '%s'", n)
		}
		ref := rn.String()
		return NewTranactionData(tim, val, ref)
	} else {
		return NewTranactionDataError("invalid Transaction node has no members (date, val and ref) '%s'", n)
	}
}
