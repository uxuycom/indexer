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
