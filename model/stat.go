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

package model

import (
	"github.com/shopspring/decimal"
	"time"
)

type ChainStatHour struct {
	ID                uint64          `gorm:"primaryKey" json:"id"`
	Chain             string          `json:"chain" gorm:"column:chain"`
	DateHour          uint32          `json:"date_hour" gorm:"column:date_hour"`
	AddressCount      uint32          `json:"address_count" gorm:"column:address_count"`
	AddressLastId     uint64          `json:"address_last_id" gorm:"column:address_last_id"`
	InscriptionsCount uint32          `json:"inscriptions_count" gorm:"column:inscriptions_count"`
	BalanceSum        decimal.Decimal `json:"balance_amount_sum" gorm:"column:balance_amount_sum;type:decimal(38,18)"`
	BalanceLastId     uint64          `json:"balance_last_id" gorm:"column:balance_last_id"`
	CreatedAt         time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

func (ChainStatHour) TableName() string {
	return "chain_stats_hour"
}
