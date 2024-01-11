package asc20

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/assert.v1"
	"open-indexer/client/xycommon"
	"open-indexer/dcache"
	"open-indexer/xylog"
	"testing"
)

func init() {
	xylog.InitLog(logrus.DebugLevel, "")
}

func TestExtractValidOrdersByExchange(t *testing.T) {
	tests := []struct {
		Name     string
		Tx       *xycommon.RpcTransaction
		Expected []*Exchange
	}{
		{
			Name: "extractValidOrdersByExchange_multiEvents",
			Tx: &xycommon.RpcTransaction{
				Events: []xycommon.RpcLog{
					{
						Address: common.HexToAddress("0x24e24277e2FF8828d5d2e278764CA258C22BD497"),
						Topics: []common.Hash{
							common.HexToHash("0xe2750d6418e3719830794d3db788aa72febcd657bcd18ed8f1facdbf61a69a9a"),
							common.HexToHash("0x00000000000000000000000023b2a8e35deea93139abc2791e895c65e1aa4c91"),
							common.HexToHash("0x00000000000000000000000081b12c1c6f6719c429648c531395c308a039b7d2"),
						},
						Data: hexutil.MustDecode("0x9e5034eacae74053ada43cfebb522c2ae6e5ffa4f3327e35cf46ae772decfbc2"),
					},
					{
						Address: common.HexToAddress("0x24e24277e2FF8828d5d2e278764CA258C22BD497"),
						Topics: []common.Hash{
							common.HexToHash("0x3efe873bf4d1c1061b9980e7aed9b564e024844522ec8c80aec160809948ef77"),
						},
						Data: hexutil.MustDecode("0x00000000000000000000000023b2a8e35deea93139abc2791e895c65e1aa4c9100000000000000000000000081b12c1c6f6719c429648c531395c308a039b7d29e5034eacae74053ada43cfebb522c2ae6e5ffa4f3327e35cf46ae772decfbc20000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000019b45a4bb00000000000000000000000000000000000000000000000000000000251d94ea00000000000000000000000000000000000000000000000000000000000000c800000000000000000000000000000000000000000000000000000000659ce65300000000000000000000000000000000000000000000000000000000000000046176617600000000000000000000000000000000000000000000000000000000"),
					},
					{
						Address: common.HexToAddress("0x24e24277e2FF8828d5d2e278764CA258C22BD497"),
						Topics: []common.Hash{
							common.HexToHash("0xe2750d6418e3719830794d3db788aa72febcd657bcd18ed8f1facdbf61a69a9a"),
							common.HexToHash("0x00000000000000000000000023b2a8e35deea93139abc2791e895c65e1aa4c91"),
							common.HexToHash("0x00000000000000000000000081b12c1c6f6719c429648c531395c308a039b7d2"),
						},
						Data: hexutil.MustDecode("0x9e5034eacae74053ada43cfebb522c2ae6e5ffa4f3327e35cf46ae772decfbc2"),
					},
					{
						Address: common.HexToAddress("0x24e24277e2FF8828d5d2e278764CA258C22BD497"),
						Topics: []common.Hash{
							common.HexToHash("0x3efe873bf4d1c1061b9980e7aed9b564e024844522ec8c80aec160809948ef77"),
						},
						Data: hexutil.MustDecode("0x00000000000000000000000023b2a8e35deea93139abc2791e895c65e1aa4c9100000000000000000000000081b12c1c6f6719c429648c531395c308a039b7d29e5034eacae74053ada43cfebb522c2ae6e5ffa4f3327e35cf46ae772decfbc20000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000019b45a4bb00000000000000000000000000000000000000000000000000000000251d94ea00000000000000000000000000000000000000000000000000000000000000c800000000000000000000000000000000000000000000000000000000659ce65300000000000000000000000000000000000000000000000000000000000000046176617600000000000000000000000000000000000000000000000000000000"),
					},
				},
			},
			Expected: []*Exchange{
				{
					Tick:   "avav",
					From:   "0x24e24277e2FF8828d5d2e278764CA258C22BD497",
					To:     "0x81b12c1c6F6719C429648C531395C308A039b7D2",
					Amount: decimal.NewFromInt(6899999931),
				},
				{
					Tick:   "avav",
					From:   "0x24e24277e2FF8828d5d2e278764CA258C22BD497",
					To:     "0x81b12c1c6F6719C429648C531395C308A039b7D2",
					Amount: decimal.NewFromInt(6899999931),
				},
			},
		},
		{
			Name: "extractValidOrdersByExchange",
			Tx: &xycommon.RpcTransaction{
				Events: []xycommon.RpcLog{
					{
						Address: common.HexToAddress("0x24e24277e2FF8828d5d2e278764CA258C22BD497"),
						Topics: []common.Hash{
							common.HexToHash("0xe2750d6418e3719830794d3db788aa72febcd657bcd18ed8f1facdbf61a69a9a"),
							common.HexToHash("0x00000000000000000000000023b2a8e35deea93139abc2791e895c65e1aa4c91"),
							common.HexToHash("0x00000000000000000000000081b12c1c6f6719c429648c531395c308a039b7d2"),
						},
						Data: hexutil.MustDecode("0x9e5034eacae74053ada43cfebb522c2ae6e5ffa4f3327e35cf46ae772decfbc2"),
					},
					{
						Address: common.HexToAddress("0x24e24277e2FF8828d5d2e278764CA258C22BD497"),
						Topics: []common.Hash{
							common.HexToHash("0x3efe873bf4d1c1061b9980e7aed9b564e024844522ec8c80aec160809948ef77"),
						},
						Data: hexutil.MustDecode("0x00000000000000000000000023b2a8e35deea93139abc2791e895c65e1aa4c9100000000000000000000000081b12c1c6f6719c429648c531395c308a039b7d29e5034eacae74053ada43cfebb522c2ae6e5ffa4f3327e35cf46ae772decfbc20000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000019b45a4bb00000000000000000000000000000000000000000000000000000000251d94ea00000000000000000000000000000000000000000000000000000000000000c800000000000000000000000000000000000000000000000000000000659ce65300000000000000000000000000000000000000000000000000000000000000046176617600000000000000000000000000000000000000000000000000000000"),
					},
				},
			},
			Expected: []*Exchange{
				{
					Tick:   "avav",
					From:   "0x24e24277e2FF8828d5d2e278764CA258C22BD497",
					To:     "0x81b12c1c6F6719C429648C531395C308A039b7D2",
					Amount: decimal.NewFromInt(6899999931),
				},
			},
		},
		{
			Name: "extractValidOrdersByExchange_failed",
			Tx: &xycommon.RpcTransaction{
				Events: []xycommon.RpcLog{
					{
						Address: common.HexToAddress("0x24e24277e2FF8828d5d2e278764CA258C22BD497"),
						Topics: []common.Hash{
							common.HexToHash("0xe2750d6418e3719830794d3db788aa72febcd657bcd18ed8f1facdbf61a69a9a"),
							common.HexToHash("0x00000000000000000000000023b2a8e35deea93139abc2791e895c65e1aa4c91"),
							common.HexToHash("0x00000000000000000000000081b12c1c6f6719c429648c531395c308a039b7d2"),
						},
						Data: hexutil.MustDecode("0x9e5034eacae74053ada43cfebb522c2ae6e5ffa4f3327e35cf46ae772decfbc2"),
					},
				},
			},
			Expected: []*Exchange{},
		},

		{
			Name: "extractValidOrdersByExchange_failed",
			Tx: &xycommon.RpcTransaction{
				Events: []xycommon.RpcLog{
					{
						Address: common.HexToAddress("0x24e24277e2FF8828d5d2e278764CA258C22BD497"),
						Topics: []common.Hash{
							common.HexToHash("0x3efe873bf4d1c1061b9980e7aed9b564e024844522ec8c80aec160809948ef77"),
						},
						Data: hexutil.MustDecode("0x00000000000000000000000023b2a8e35deea93139abc2791e895c65e1aa4c9100000000000000000000000081b12c1c6f6719c429648c531395c308a039b7d29e5034eacae74053ada43cfebb522c2ae6e5ffa4f3327e35cf46ae772decfbc20000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000019b45a4bb00000000000000000000000000000000000000000000000000000000251d94ea00000000000000000000000000000000000000000000000000000000000000c800000000000000000000000000000000000000000000000000000000659ce65300000000000000000000000000000000000000000000000000000000000000046176617600000000000000000000000000000000000000000000000000000000"),
					},
				},
			},
			Expected: []*Exchange{},
		},
	}

	cache := dcache.NewManager(nil, "avax")
	protocol := NewProtocol(cache)
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := protocol.extractValidOrdersByExchange(test.Tx)
			assert.Equal(t, len(result), len(test.Expected))
			if len(result) == len(test.Expected) {
				for idx, item := range result {
					assert.Equal(t, item.Tick, test.Expected[idx].Tick)
					assert.Equal(t, item.From, test.Expected[idx].From)
					assert.Equal(t, item.To, test.Expected[idx].To)
					assert.Equal(t, item.Amount, test.Expected[idx].Amount)
				}
			}
		})
	}
}

func TestExtractValidOrdersByTransfer(t *testing.T) {
	tests := []struct {
		Name     string
		Tx       *xycommon.RpcTransaction
		Expected []*Exchange
		Tickers  []string
	}{
		{
			Name:    "extractValidOrdersByTransfers",
			Tickers: []string{"avax"},
			Tx: &xycommon.RpcTransaction{
				Events: []xycommon.RpcLog{
					{
						Address: common.BytesToAddress([]byte("0x24e24277e2FF8828d5d2e278764CA258C22BD497")),
						Topics: []common.Hash{
							common.HexToHash("0x8cdf9e10a7b20e7a9c4e778fc3eb28f2766e438a9856a62eac39fbd2be98cbc2"),
							common.HexToHash("0x000000000000000000000000c37c800260cd7b766bf870d930a696b98259c546"),
							common.HexToHash("0x000000000000000000000000117b15af63e1d533cc5bac7333f3cc8f8cc2696d"),
							common.HexToHash("0x51ae1b9bb3103c91be3c12db9f97165657aee56ce412966fd68b8715b0481595"),
						},
						Data: hexutil.MustDecode("0x00000000000000000000000000000000000000000000000000000000000001f4"),
					},
				},
			},
			Expected: []*Exchange{
				{
					Tick:   "avax",
					From:   "0xC37C800260cd7B766bf870d930a696b98259C546",
					To:     "0x117b15af63E1D533cc5baC7333f3Cc8f8Cc2696d",
					Amount: decimal.NewFromInt(500),
				},
			},
		},
	}

	cache := dcache.NewManager(nil, "avax")
	cache.Inscription = dcache.NewInscription()
	protocol := NewProtocol(cache)
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			for _, tick := range test.Tickers {
				cache.Inscription.Create("asc-20", tick, &dcache.Tick{})
			}

			result := protocol.extractValidOrdersByTransfer(test.Tx)
			assert.Equal(t, len(result), len(test.Expected))
			if len(result) == len(test.Expected) {
				for idx, item := range result {
					assert.Equal(t, item.Tick, test.Expected[idx].Tick)
					assert.Equal(t, item.From, test.Expected[idx].From)
					assert.Equal(t, item.To, test.Expected[idx].To)
					assert.Equal(t, item.Amount, test.Expected[idx].Amount)
				}
			}
		})
	}
}
