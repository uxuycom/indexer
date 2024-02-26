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

package utils

import (
	"fmt"
	"github.com/uxuycom/indexer/xylog"
	"strconv"
	"time"
)

var (
	dateFormat     = "20060102"
	dateLineFormat = "2006-01-02"
	hourFormat     = "2006010215"
	hourLineFormat = "2006-01-02-15"
	timeFormat     = "20060102 15:23:26"
	timeLineFormat = "2006-01-02 15:23:26"
)

func BeforeYesterdayHour() time.Time {
	yesterday := time.Now().Add(-24 * time.Hour)
	return Hour(yesterday.Add(-24 * time.Hour))
}

func YesterdayHour() time.Time {
	return Hour(time.Now().Add(-24 * time.Hour))
}

func Hour(tm time.Time) time.Time {
	return tm.Truncate(time.Hour)
}

func TimeHourInt(tm time.Time) uint64 {
	format := tm.Format(hourFormat)
	tmInt, err := strconv.ParseUint(format, 10, 32)
	if err != nil {
		xylog.Logger.Errorf("TimeHourInt err!, err = %v ", err)
	}
	return tmInt
}

func All() time.Time {
	// get 24H chain stat from chain_stats_hour
	now := time.Now().Truncate(time.Hour)
	yesterday := now.Add(-24 * time.Hour).Truncate(time.Hour)
	dayBeforeYesterday := yesterday.Add(-24 * time.Hour).Truncate(time.Hour)
	_ = fmt.Sprintf("dayBeforeYesterday:%v", dayBeforeYesterday)

	nowFormat := now.Format("2006010215")
	nowUint, _ := strconv.ParseUint(nowFormat, 10, 32)
	yesterdayFormat := yesterday.Format("2006010215")
	yesterdayUint, _ := strconv.ParseUint(yesterdayFormat, 10, 32)
	dayBeforeYesterdayFormat := dayBeforeYesterday.Format("2006010215")

	fmt.Printf("%v  %v %v ", nowUint, yesterdayUint, dayBeforeYesterdayFormat)

	return dayBeforeYesterday
}
