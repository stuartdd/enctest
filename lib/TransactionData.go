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
)

type TranactionData struct {
	dateTime time.Time
	value    float64
	ref      string
	err      error

	lineValue float64
}

type TranactionDataList struct {
	Node         parser.NodeC      // The !tx node
	InitialValue float64           // initial value.
	ClosingValue float64           // initial value -+ all transactions
	Data         []*TranactionData // Each transaction
}

// !tx node. List of all transactions.
// Sorted by datetime.
func NewTranactionDataList(n parser.NodeC, initialValue float64) *TranactionDataList {
	d := make([]*TranactionData, n.Len())
	v := initialValue
	for i, ni := range n.GetValues() {
		d[i] = NewTranactionDataFromNode(ni)
	}
	sort.Slice(d, func(i, j int) bool {
		return d[i].dateTime.UnixMilli() < d[j].dateTime.UnixMilli()
	})
	for _, ni := range d {
		ni.SetLineValue(v - ni.Value())
		v = ni.LineValue()
	}
	return &TranactionDataList{Node: n, InitialValue: initialValue, ClosingValue: v, Data: d}
}

func (t *TranactionDataList) String() string {
	return fmt.Sprintf("Initial value %9.2f.    Final value %9.2f", t.InitialValue, t.ClosingValue)
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
