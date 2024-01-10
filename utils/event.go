package utils

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"reflect"
	"unicode"
)

type EventLog struct {
	Address common.Address `json:"address" gencodec:"required"`
	Topics  []common.Hash  `json:"topics" gencodec:"required"`
	Data    hexutil.Bytes  `json:"data" gencodec:"required"`
}

func ParseEventToMap(parsedABI abi.ABI, eventLog EventLog, output map[string]interface{}) (eventName string, err error) {
	if output == nil || reflect.TypeOf(output).Kind() != reflect.Ptr {
		return "", fmt.Errorf("output must be a pointer")
	}

	if len(eventLog.Topics) < 1 {
		return "", fmt.Errorf("log topics = 0")
	}

	findEvent, err := parsedABI.EventByID(eventLog.Topics[0])
	if err != nil {
		return "", fmt.Errorf("call EventByID error[%v]", err)
	}

	if findEvent == nil {
		return "", fmt.Errorf("event[%s] not found", eventLog.Topics[0])
	}

	eventName = findEvent.Name
	if unicode.IsDigit(rune(findEvent.Name[len(findEvent.Name)-1])) {
		eventName = findEvent.Name[:len(findEvent.Name)-1]
	}

	err = parsedABI.UnpackIntoMap(output, findEvent.Name, eventLog.Data)
	if err != nil {
		return "", fmt.Errorf("UnpackIntoInterface error[%v]]", err)
	}

	// build args
	args := make([]abi.Argument, 0, len(findEvent.Inputs))
	for _, arg := range findEvent.Inputs {
		if arg.Indexed {
			args = append(args, arg)
		}
	}

	if len(args) <= 0 {
		return
	}

	//build topics
	topics := eventLog.Topics[1:]
	err = abi.ParseTopicsIntoMap(output, args, topics)
	if err != nil {
		return "", fmt.Errorf("failed to parse topics into TransferEvent: %v", err)
	}
	return
}

func ParseEventToStruct(parsedABI abi.ABI, eventLog EventLog, output interface{}) (eventName string, err error) {
	if output == nil || reflect.TypeOf(output).Kind() != reflect.Ptr {
		return "", fmt.Errorf("output must be a pointer")
	}

	if len(eventLog.Topics) < 1 {
		return "", fmt.Errorf("log topics = 0")
	}

	findEvent, err := parsedABI.EventByID(eventLog.Topics[0])
	if err != nil {
		return "", fmt.Errorf("call EventByID error[%v]", err)
	}

	if findEvent == nil {
		return "", fmt.Errorf("event[%s] not found", eventLog.Topics[0])
	}

	eventName = findEvent.Name
	if unicode.IsDigit(rune(findEvent.Name[len(findEvent.Name)-1])) {
		eventName = findEvent.Name[:len(findEvent.Name)-1]
	}

	err = parsedABI.UnpackIntoInterface(output, findEvent.Name, eventLog.Data)
	if err != nil {
		return "", fmt.Errorf("UnpackIntoInterface error[%v]]", err)
	}

	// build args
	args := make([]abi.Argument, 0, len(findEvent.Inputs))
	for _, arg := range findEvent.Inputs {
		if arg.Indexed {
			args = append(args, arg)
		}
	}

	if len(args) <= 0 {
		return
	}

	//build topics
	topics := eventLog.Topics[1:]
	err = abi.ParseTopics(output, args, topics)
	if err != nil {
		return "", fmt.Errorf("failed to parse topics into TransferEvent: %v", err)
	}
	return
}
