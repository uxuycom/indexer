package btc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alitto/pond"
	"github.com/uxuycom/indexer/xylog"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
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

func (c *OrdClient) doCallContext(ctx context.Context, path string, out interface{}) error {
	startTs := time.Now()
	defer func() {
		xylog.Logger.Infof("call ord api[%s] cost[%v]", path, time.Since(startTs))
	}()

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
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode == http.StatusNotFound {
		return nil
	}

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

func (c *OrdClient) callContext(ctx context.Context, path string, out interface{}) (err error) {
	ts := time.Millisecond * 100
	for retry := 0; retry < 5; retry++ {
		err = c.doCallContext(ctx, path, out)
		if err == nil {
			return nil
		}
		<-time.After(ts * time.Duration(retry))
	}
	return err
}

func (c *OrdClient) blockInscriptionsByPage(ctx context.Context, blockNum int64, page int) (ret BlockInscriptions, err error) {
	path := fmt.Sprintf("inscriptions/block/%d/%d", blockNum, page)
	err = c.callContext(ctx, path, &ret)
	return
}

type Inscription struct {
	ID      string
	Meta    InscriptionMeta
	Content string
}

func (c *OrdClient) BlockBRC20Inscriptions(ctx context.Context, blockNum int64) (map[string]Inscription, error) {
	ids, err := c.BlockInscriptionIDs(ctx, blockNum)
	if err != nil {
		return nil, fmt.Errorf("call BlockInscriptionIDs error: %v", err)
	}

	// get inscriptions meta
	metas, err := c.InscriptionMetaByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("call InscriptionMetaByIDs error: %v", err)
	}

	validIds := make([]string, 0, len(ids))
	for _, meta := range metas {
		// meta.ContentType must contains "text/plain" or "application/json".
		if !strings.Contains(meta.ContentType, "text/plain") && !strings.Contains(meta.ContentType, "application/json") {
			continue
		}
		validIds = append(validIds, meta.InscriptionID)
	}

	// get inscriptions content
	contents, err := c.InscriptionContentByIDs(ctx, validIds)
	if err != nil {
		return nil, fmt.Errorf("call InscriptionContentByIDs error: %v", err)
	}

	result := make(map[string]Inscription, len(validIds))
	for id, content := range contents {
		txId := strings.Split(id, "i")[0]

		if _, ok := result[txId]; ok {
			xylog.Logger.Fatalf("txId[%s] has more than one inscription", txId)
		}

		result[txId] = Inscription{
			ID:      id,
			Meta:    metas[id],
			Content: content,
		}
	}
	return result, nil
}

func (c *OrdClient) BlockInscriptionIDs(ctx context.Context, blockNum int64) ([]string, error) {
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

func (c *OrdClient) InscriptionContentByIDs(ctx context.Context, insIDs []string) (map[string]string, error) {
	pool := pond.New(100, 0, pond.MinWorkers(100))
	contentsMap := &sync.Map{}
	for _, insID := range insIDs {
		id := insID
		pool.Submit(func() {
			content, err := c.InscriptionContentByID(ctx, id)
			if err != nil {
				xylog.Logger.Errorf("get inscription content err[%v], id[%s]", err, id)
				return
			}
			contentsMap.Store(id, content)
		})
	}
	pool.StopAndWait()

	result := make(map[string]string, len(insIDs))
	for _, insID := range insIDs {
		v, ok := contentsMap.Load(insID)
		if !ok {
			return nil, fmt.Errorf("get inscription content nil, id[%s]", insID)
		}
		result[insID] = v.(string)
	}
	return result, nil
}

func (c *OrdClient) InscriptionContentByID(ctx context.Context, id string) (ret string, err error) {
	path := fmt.Sprintf("content/%s", id)
	err = c.callContext(ctx, path, &ret)
	return
}

func (c *OrdClient) InscriptionMetaByIDs(ctx context.Context, insIDs []string) (map[string]InscriptionMeta, error) {
	pool := pond.New(100, 0, pond.MinWorkers(100))
	inscriptionsMap := &sync.Map{}
	for _, insID := range insIDs {
		id := insID
		pool.Submit(func() {
			inscription, err := c.InscriptionMetaByID(ctx, id)
			if err != nil {
				xylog.Logger.Errorf("get inscription meta err[%v], id[%s]", err, id)
				return
			}
			inscriptionsMap.Store(id, inscription)
		})
	}
	pool.StopAndWait()

	result := make(map[string]InscriptionMeta, len(insIDs))
	for _, insID := range insIDs {
		v, ok := inscriptionsMap.Load(insID)
		if !ok {
			return nil, fmt.Errorf("get inscription meta nil, id[%s]", insID)
		}
		result[insID] = v.(InscriptionMeta)
	}
	return result, nil
}

type InscriptionMeta struct {
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

func (c *OrdClient) InscriptionMetaByID(ctx context.Context, id string) (ret InscriptionMeta, err error) {
	path := fmt.Sprintf("inscription/%s", id)
	err = c.callContext(ctx, path, &ret)
	return
}

type InscriptionOutput struct {
	Value        int64    `json:"value"`
	ScriptPubkey string   `json:"script_pubkey"`
	Address      string   `json:"address"`
	Transaction  string   `json:"transaction"`
	SatRanges    []int    `json:"sat_ranges,omitempty"`
	Inscriptions []string `json:"inscriptions"`
}

func (c *OrdClient) InscriptionOutput(ctx context.Context, txID string, index uint32) (ret InscriptionOutput, err error) {
	path := fmt.Sprintf("output/%s:%d", txID, index)
	err = c.callContext(ctx, path, &ret)
	return
}
