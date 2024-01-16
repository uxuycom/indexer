package cache_store

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestCacheStore_Set(t *testing.T) {
	type fields struct {
		MaxCapacity int64
		Duration    uint32
	}
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "accuracy test",
			fields: fields{
				MaxCapacity: 10,
				Duration:    7,
			},
			args: args{
				key:   "ins",
				value: "{\"jsonrpc\":\"2.0\",\"result\":{\"inscriptions\":[{\"chain\":\"avalanche\",\"protocol\":\"asc-20\",\"tick\":\"crazy dog\",\"deploy_by\":\"0xB197D5Bbd55F7867964D09b96928926d29DD0411\",\"deploy_hash\":\"0x25ec9058fa19b26de3e2ca4596555f72de3987cc9b893decc779197e5683c7aa\",\"total_supply\":\"1200000\",\"minted_percent\":\"1.0000\",\"limit_per_mint\":\"1000\",\"holders\":6,\"transfer_type\":0,\"status\":2,\"minted\":\"1200000\",\"tx_cnt\":1203,\"created_at\":1705071307},{\"chain\":\"avalanche\",\"protocol\":\"asc-20\",\"tick\":\"%5Cu76D8%5Cu53E4\",\"deploy_by\":\"0xB197D5Bbd55F7867964D09b96928926d29DD0411\",\"deploy_hash\":\"0xf57ffaa81c5d4ea78249130528d4e27c7ebe1238f6d3c059754c74f942f0e099\",\"total_supply\":\"1200000\",\"minted_percent\":\"0.0000\",\"limit_per_mint\":\"1000\",\"holders\":0,\"transfer_type\":0,\"status\":1,\"minted\":\"0\",\"tx_cnt\":1,\"created_at\":1705071300},{\"chain\":\"avalanche\",\"protocol\":\"asc-20\",\"tick\":\"MadDog\",\"deploy_by\":\"0xB197D5Bbd55F7867964D09b96928926d29DD0411\",\"deploy_hash\":\"0x16daffbcc3e201ae5304c2b7add859720736f28347c7c0d52063c4d487c66558\",\"total_supply\":\"1200000\",\"minted_percent\":\"0.0000\",\"limit_per_mint\":\"1000\",\"holders\":0,\"transfer_type\":0,\"status\":1,\"minted\":\"0\",\"tx_cnt\":1,\"created_at\":1705071299},{\"chain\":\"avalanche\",\"protocol\":\"asc-20\",\"tick\":\"%5Cu8001%5Cu5B50\",\"deploy_by\":\"0x871691BA63278b5828E875C6883a32d2Bbe213f5\",\"deploy_hash\":\"0x68bebbc38d4d21937afa893f06b1db174962c0b8e1e3ff01cac157339e15acb8\",\"total_supply\":\"50000\",\"minted_percent\":\"0.0000\",\"limit_per_mint\":\"49\",\"holders\":0,\"transfer_type\":0,\"status\":1,\"minted\":\"0\",\"tx_cnt\":1,\"created_at\":1705071284},{\"chain\":\"avalanche\",\"protocol\":\"asc-20\",\"tick\":\"MJ\",\"deploy_by\":\"0xaeaD3706acc49a122f8AfF78F598F207c9D8cAC5\",\"deploy_hash\":\"0xedcf1d3904ad4f0109f6bd78d616b1a2d07e922eadcdf06e61938a67a5875f87\",\"total_supply\":\"2100000000\",\"minted_percent\":\"0.0000\",\"limit_per_mint\":\"100000\",\"holders\":1,\"transfer_type\":0,\"status\":1,\"minted\":\"100000\",\"tx_cnt\":2,\"created_at\":1705071274},{\"chain\":\"avalanche\",\"protocol\":\"asc-20\",\"tick\":\"Snoopy\",\"deploy_by\":\"0x7dAa8a69502AEF104cb593be2F4804C3E7C5eb89\",\"deploy_hash\":\"0x54e6435b63e24299560e04324aa15261070af346a466890d2735b1d711072363\",\"total_supply\":\"21000000\",\"minted_percent\":\"0.0001\",\"limit_per_mint\":\"1000\",\"holders\":1,\"transfer_type\":0,\"status\":1,\"minted\":\"3000\",\"tx_cnt\":4,\"created_at\":1705071142},{\"chain\":\"avalanche\",\"protocol\":\"asc-20\",\"tick\":\"John Wick\",\"deploy_by\":\"0x7dAa8a69502AEF104cb593be2F4804C3E7C5eb89\",\"deploy_hash\":\"0x172ae1ec7a37c118b153d60b46a23da4c4262c8d03935fe14035e76ce0c90216\",\"total_supply\":\"210000000\",\"minted_percent\":\"0.0038\",\"limit_per_mint\":\"100000\",\"holders\":1,\"transfer_type\":0,\"status\":1,\"minted\":\"800000\",\"tx_cnt\":9,\"created_at\":1705071140},{\"chain\":\"avalanche\",\"protocol\":\"asc-20\",\"tick\":\"Superman\",\"deploy_by\":\"0x7dAa8a69502AEF104cb593be2F4804C3E7C5eb89\",\"deploy_hash\":\"0x8a6c91101ef2bf26134a25d1364c805c5d5ce818fa7abf6a9f70c32a4ea1ade4\",\"total_supply\":\"1000\",\"minted_percent\":\"0.0120\",\"limit_per_mint\":\"1\",\"holders\":2,\"transfer_type\":0,\"status\":1,\"minted\":\"12\",\"tx_cnt\":14,\"created_at\":1705071140},{\"chain\":\"avalanche\",\"protocol\":\"asc-20\",\"tick\":\"Rambo\",\"deploy_by\":\"0x7dAa8a69502AEF104cb593be2F4804C3E7C5eb89\",\"deploy_hash\":\"0xc3b4bfe8e0dbc908e6ff44d1f55f1720cc13f38f630510ce6d1015eed3a06b07\",\"total_supply\":\"21000000000\",\"minted_percent\":\"0.0004\",\"limit_per_mint\":\"1000000\",\"holders\":1,\"transfer_type\":0,\"status\":1,\"minted\":\"9000000\",\"tx_cnt\":10,\"created_at\":1705071137},{\"chain\":\"avalanche\",\"protocol\":\"asc-20\",\"tick\":\"Otto\",\"deploy_by\":\"0x7dAa8a69502AEF104cb593be2F4804C3E7C5eb89\",\"deploy_hash\":\"0x430c1d6916b3c308027cde95bb0d6e6c39220e2c87f4416a2b473cb07f4ddbfd\",\"total_supply\":\"21000000\",\"minted_percent\":\"0.0000\",\"limit_per_mint\":\"1000\",\"holders\":0,\"transfer_type\":0,\"status\":1,\"minted\":\"0\",\"tx_cnt\":1,\"created_at\":1705071074}],\"total\":6361,\"limit\":10,\"offset\":0},\"id\":1}",
			},
		},
	}
	randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCacheStore(tt.fields.MaxCapacity, tt.fields.Duration)
			go m.Clear()
			for i := 0; i < 10000; i++ {
				if i%1000 == 0 {
					time.Sleep(1 * time.Second)
				}
				key := fmt.Sprintf("%s_%v", tt.args.key, randomGenerator.Intn(10000))
				m.Set(key, tt.args.value)
			}

			for i := 0; i < 10000; i++ {
				key := fmt.Sprintf("%s_%v", tt.args.key, randomGenerator.Intn(10000))
				if _, ok := m.Get(key); ok {
					fmt.Printf("hit cache. key=%s\n", key)
				}
			}
		})
	}
}
