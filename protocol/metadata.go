package protocol

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"open-indexer/client/xycommon"
	"open-indexer/devents"
	"open-indexer/model"
	"open-indexer/protocol/avax/asc20"
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
		if tx.Input[:4] == asc20.ExchangeMethodID {
			return asc20.ParseMetaDataByEventLogs(chainName, tx)
		}
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
	return nil, nil
}
