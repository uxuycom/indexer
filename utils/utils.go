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

package utils

import (
	"encoding/hex"
	"fmt"
	"github.com/wealdtech/go-merkletree/keccak256"
	"math/big"
	"strconv"
	"strings"
)

func HexToUint64(hex string) uint64 {
	num, ok := new(big.Int).SetString(hex[2:], 16)
	if ok {
		return num.Uint64()
	} else {
		return 0
	}
}

func ParseInt64(str string) int64 {
	if strings.Contains(str, ".") {
		str = strings.Split(str, ".")[0]
	}
	rst, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	} else {
		return rst
	}
}

func ConvetStr(number string) (*big.Int, error) {
	if number != "" {
		max_big, is_ok := new(big.Int).SetString(number, 10)
		if !is_ok {
			return big.NewInt(0), fmt.Errorf("number error")
		}
		return max_big, nil
	}
	return big.NewInt(0), nil
}

func Keccak256(str string) string {
	h := keccak256.New()
	bytes := h.Hash([]byte(str))
	return hex.EncodeToString(bytes)
}
