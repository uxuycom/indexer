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

package types

import (
	"github.com/uxuycom/indexer/client/xycommon"
	"github.com/uxuycom/indexer/devents"
	"github.com/uxuycom/indexer/xyerrors"
)

type IProtocol interface {
	Parse(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, md *devents.MetaData) ([]*devents.TxResult, *xyerrors.InsError)
}

const (
	BRC20Protocol = "brc-20"
	ASC20Protocol = "asc-20"
	BSC20Protocol = "bsc-20"
	PRC20Protocol = "prc-20"
	ERC20Protocol = "erc-20"

	DefaultMaxDataLength = 256
)

// DefaultMaxDataLengthMap max data length
var DefaultMaxDataLengthMap = map[string]int{
	BRC20Protocol: 256,
	ASC20Protocol: 256,
	BSC20Protocol: 256,
	PRC20Protocol: 256,
	ERC20Protocol: 256,
}
