// Copyright (c) 2023-2024 The UXUY Developer Team
// License:
// MIT License

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE

package brc20

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/xyerrors"
	"strings"
)

const MaxPrecision = 18

// NewDecimalFromString
// Numeric fields are not stripped/trimmed. "dec" field must have only digits, other numeric fields may have a single dot(".") for decimal representation (+,- etc. are not accepted). Decimal fields cannot start or end with dot (e.g. ".99" and "99." are invalid).
// https://layer1.gitbook.io/layer1-foundation/protocols/brc-20/indexing#rules
///****

func NewDecimalFromString(s string) (decimal.Decimal, int, error) {
	if s == "" {
		return decimal.Zero, 0, nil
	}

	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return decimal.Zero, 0, fmt.Errorf("invalid decimal format, s[%s]", s)
	}

	if len(parts) == 1 {
		d, err := decimal.NewFromString(s)
		if err != nil {
			return decimal.Zero, 0, err
		}
		return d, 0, nil
	}

	intPartStr := parts[0]
	if intPartStr == "" || intPartStr[0] == '+' || intPartStr[0] == '-' {
		return decimal.Zero, 0, fmt.Errorf("invalid decimal int part prefix, s[%s]", s)
	}

	fractionalPartStr := parts[1]
	if fractionalPartStr == "" || fractionalPartStr[0] == '-' || fractionalPartStr[0] == '+' {
		return decimal.Zero, 0, fmt.Errorf("invalid decimal fractioinal part suffix, s[%s]", s)
	}

	if len(fractionalPartStr) > MaxPrecision {
		return decimal.Zero, 0, fmt.Errorf("decimal exceeds maximum precision: %s", s)
	}

	d, err := decimal.NewFromString(s)
	return d, len(fractionalPartStr), err
}

type Amt struct {
	Amount string `json:"amt"`
}

func ParseAmountParam(data string) (decimal.Decimal, int, *xyerrors.InsError) {
	params := &Amt{}
	err := json.Unmarshal([]byte(data), params)
	if err != nil {
		return decimal.Zero, 0, xyerrors.NewInsError(-10, fmt.Sprintf("data json deocde err:%v, data[%s]", err, data))
	}
	amount, precision, err := NewDecimalFromString(params.Amount)
	if err != nil {
		return decimal.Zero, 0, xyerrors.NewInsError(-11, fmt.Sprintf("amount format error, amount[%s]", params.Amount))
	}
	return amount, precision, nil
}
