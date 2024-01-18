package btc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type OrdClient struct {
	endpoint string
	client   *http.Client
}

func NewOrdClient(endpoint string) *OrdClient {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
	return &OrdClient{
		endpoint: strings.TrimRight(strings.TrimSpace(endpoint), "/"),
		client:   client,
	}
}

type BlockInscriptions struct {
	Inscriptions []string `json:"inscriptions"`
	More         bool     `json:"more"`
	PageIndex    int      `json:"page_index"`
}

func (c *OrdClient) callContext(ctx context.Context, path string, out interface{}) error {
	// check out whether is a pointer
	if reflect.TypeOf(out).Kind() != reflect.Ptr {
		return fmt.Errorf("out should be a pointer")
	}

	uri := fmt.Sprintf("%s/%s", c.endpoint, strings.TrimLeft(path, "/"))
	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// set headers
	req.Header.Set("Accept", "application/json")

	response, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	if len(data) == 0 {
		return nil
	}

	// check if out is a []byte
	if reflect.TypeOf(out).Elem().Kind() == reflect.Slice {
		if reflect.TypeOf(out).Elem().Elem().Kind() == reflect.Uint8 {
			reflect.ValueOf(out).Elem().SetBytes(data)
			return nil
		}
	}

	// check if out is a string
	if reflect.TypeOf(out).Elem().Kind() == reflect.String {
		reflect.ValueOf(out).Elem().SetString(string(data))
		return nil
	}

	err = json.Unmarshal(data, out)
	if err != nil {
		return fmt.Errorf("error parsing response body[%s], err[%v]", string(data), err)
	}
	return nil
}

func (c *OrdClient) blockInscriptionsByPage(ctx context.Context, blockNum int64, page int) (ret BlockInscriptions, err error) {
	path := fmt.Sprintf("inscriptions/block/%d/%d", blockNum, page)
	err = c.callContext(ctx, path, &ret)
	return
}

func (c *OrdClient) BlockInscriptions(ctx context.Context, blockNum int64) ([]string, error) {
	ids := make([]string, 0, 1000)
	page := 0
	for {
		ret, err := c.blockInscriptionsByPage(ctx, blockNum, page)
		if err != nil {
			return nil, fmt.Errorf("call blockInscriptionsByPage error: %v", err)
		}

		if len(ret.Inscriptions) > 0 {
			ids = append(ids, ret.Inscriptions...)
		}

		if !ret.More {
			break
		}
		page++
	}
	return ids, nil
}

func (c *OrdClient) InscriptionContent(ctx context.Context, id string) (ret string, err error) {
	path := fmt.Sprintf("content/%s", id)
	err = c.callContext(ctx, path, &ret)
	return
}

type Inscription struct {
	Address           string `json:"address"`
	ContentLength     int    `json:"content_length"`
	ContentType       string `json:"content_type"`
	GenesisFee        int    `json:"genesis_fee"`
	GenesisHeight     int64  `json:"genesis_height"`
	InscriptionID     string `json:"inscription_id"`
	InscriptionNumber int64  `json:"inscription_number"`
	Next              string `json:"next"`
	OutputValue       int64  `json:"output_value"`
	SatPoint          string `json:"satpoint"`
	Timestamp         int64  `json:"timestamp"`
}

func (c *OrdClient) Inscription(ctx context.Context, id string) (ret Inscription, err error) {
	path := fmt.Sprintf("inscription/%s", id)
	err = c.callContext(ctx, path, &ret)
	return
}
