package storage

import (
	"errors"
	"gorm.io/gorm"
	"open-indexer/model"
	"open-indexer/utils"
	"testing"
)

func TestNewSqliteClient(t *testing.T) {
	tests := []struct {
		name     string
		cfg      utils.DatabaseConfig
		expected *DBClient
	}{
		{
			name: "success",
			cfg: utils.DatabaseConfig{
				Dsn: "gauss-indexer.db",
			},
			expected: &DBClient{
				SqlDB: nil,
			},
		},
		{
			name: "error",
			cfg: utils.DatabaseConfig{
				Dsn: "",
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, _ := NewSqliteClient(&test.cfg, nil)
			if actual != test.expected {
				t.Errorf("NewSqliteClient(%v) = %v, want %v", test.cfg, actual, test.expected)
			}
		})
	}
}

func TestDBClient_Stop(t *testing.T) {
	tests := []struct {
		name     string
		client   *DBClient
		expected error
	}{
		{
			name: "success",
			client: &DBClient{
				SqlDB: nil,
			},
			expected: nil,
		},
		{
			name: "error",
			client: &DBClient{
				SqlDB: &gorm.DB{},
			},
			expected: errors.New("sql: database is closed"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.client.SqlDB != nil {
				t.Errorf("DBClient.Stop() did not close the database connection")
			}
		})
	}
}

func TestDBClient_UpdateBlock(t *testing.T) {
	tests := []struct {
		name      string
		client    *DBClient
		height    int64
		blockHash string
		expected  error
	}{
		{
			name: "success",
			client: &DBClient{
				SqlDB: &gorm.DB{},
			},
			height:    1,
			blockHash: "0x1234567890abcdef",
			expected:  nil,
		},
		{
			name: "error",
			client: &DBClient{
				SqlDB: nil,
			},
			height:    1,
			blockHash: "0x1234567890abcdef",
			expected:  errors.New("sql: database is closed"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.client.SaveLastBlock(nil, &model.BlockStatus{})
			if err != test.expected {
				t.Errorf("DBClient.UpdateBlock(%v, %v) = %v, want %v", test.height, test.blockHash, err, test.expected)
			}
		})
	}
}
