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

package protocol

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/uxuycom/indexer/client/xycommon"
	"github.com/uxuycom/indexer/devents"
	"github.com/uxuycom/indexer/model"
	"github.com/uxuycom/indexer/protocol/avax/asc20"
	"github.com/uxuycom/indexer/protocol/btc/brc20"
	"strings"
)

var EVMValidContentTypes = map[string]struct{}{
	"":                 {},
	"text/plain":       {},
	"application/json": {},
}

func ParseMetaData(chainName string, tx *xycommon.RpcTransaction) (*devents.MetaData, error) {
	switch chainName {
	case model.ChainBTC:
		return ParseBTCMetaData(chainName, tx)
	}

	// MethodID: 0xd9b3d6d0
	if chainName == model.ChainAVAX && len(tx.Events) > 0 {
		return asc20.ParseMetaDataByEventLogs(chainName, tx)
	}
	return ParseEVMMetaData(chainName, tx.Input)
}

func ParseEVMMetaData(chain string, inputData string) (*devents.MetaData, error) {
	// 0x prefix checking
	if !strings.HasPrefix(inputData, "0x") {
		return nil, fmt.Errorf("input 0x prefix checking failed")
	}

	bytes, err := hex.DecodeString(inputData[2:])
	if err != nil {
		return nil, fmt.Errorf("input hex data decode err:%v", err)
	}

	// try json format data
	input := string(bytes)
	dataPrefixIdx := strings.Index(input, ",")
	if dataPrefixIdx == -1 {
		return nil, fmt.Errorf("data seprator index failed")
	}

	// max length limit
	if len(input) > 256 {
		return nil, fmt.Errorf("data character size[%d] > 256", len(input))
	}

	//set parse content types
	contentType := ""
	if dataPrefixIdx > 5 {
		contentType = input[5:dataPrefixIdx]
	}
	contentType = strings.ToLower(contentType)
	if _, ok := EVMValidContentTypes[contentType]; !ok {
		return nil, fmt.Errorf("tx content-type invalid & filtered, ct:%s", contentType)
	}

	data := input[dataPrefixIdx+1:]
	proto := &devents.MetaData{}
	if err := json.Unmarshal([]byte(data), proto); err != nil {
		return nil, fmt.Errorf("tx input data parsed failed, data[%s], err[%v]", data, err)
	}

	// trim prefix / suffix spaces & case insensitive
	proto.Protocol = strings.ToLower(strings.TrimSpace(proto.Protocol))
	proto.Operate = strings.ToLower(strings.TrimSpace(proto.Operate))
	proto.Tick = strings.ToLower(strings.TrimSpace(proto.Tick))

	// data checking
	if proto.Protocol == "" || proto.Tick == "" {
		return nil, fmt.Errorf("tx input data protocol / tick empty, data[%s]", data)
	}
	proto.Chain = chain
	proto.Data = data
	return proto, nil
}

func ParseBTCMetaData(chain string, tx *xycommon.RpcTransaction) (*devents.MetaData, error) {
	return brc20.ParseMetaData(tx)
}
