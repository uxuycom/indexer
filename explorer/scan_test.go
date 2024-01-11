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

package explorer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/uxuycom/indexer/xylog"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestTransactionDataByHash(t *testing.T) {

	//txhash := "0x8166ff37f1fb6b1d2cdb6b8759ca4c790b1ea2ea14bee22ffc0434b81c3d2050"
	url := "https://1rpc.io/avax/c"
	method := "POST"

	payload := strings.NewReader(`{"method":"eth_getTransactionByHash","params":["0x8166ff37f1fb6b1d2cdb6b8759ca4c790b1ea2ea14bee22ffc0434b81c3d2050"],"id":1,"jsonrpc":"2.0"}`)
	//payload := strings.NewReader(`{"method":"eth_getTransactionByHash","params":["0xcb85b1bd78e39eedd4205ef982b8cb6418f3a8452deceaf6ab451fddee3b58bf"],"id":1,"jsonrpc":"2.0"}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		xylog.Logger.Error(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		if xylog.Logger != nil {
			xylog.Logger.Error(err)
		}
		return
	}
	if res != nil {
		defer func() {
			_ = res.Body.Close()
		}()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		xylog.Logger.Error(err)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		//fmt.Println("Error:", err)
		xylog.Logger.Error(err)
		return
	}

	subData := data["result"].(map[string]interface{})
	fmt.Println("Data>>> :", subData["input"])

	var hexString = fmt.Sprintf("%s", subData["input"])
	// Decode hex string to byte slice
	hexBytes, err := hex.DecodeString(hexString[2:])
	if err != nil {
		fmt.Println("Error decoding hex string:", err)
		return
	}

	// Convert byte slice to string
	resultString := string(hexBytes)

	fmt.Println("Original Hex String:", hexString)
	fmt.Println("Converted String:", resultString)
	//fmt.Println("input:", data["input"])
}

func Test(t *testing.T) {

	url := "https://1rpc.io/avax/c"
	method := "POST"

	payload := strings.NewReader(`{"method":"eth_blockNumber","params":[],"id":1,"jsonrpc":"2.0"}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
