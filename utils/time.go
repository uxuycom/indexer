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
	"github.com/uxuycom/indexer/xylog"
	"strconv"
	"time"
)

var (
	dateFormat     = "20060102"
	dateLineFormat = "2006-01-02"
	hourFormat     = "2006010215"
	hourLineFormat = "2006-01-02-15"
	timeFormat     = "20060102 15:04:05"
	timeLineFormat = "2006-01-02 15:04:05"
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

// FirstDayOfMonth the first day of the month
func FirstDayOfMonth() int64 {
	formattedTime := time.Now().Format("200601")
	beginDateStr := formattedTime + "01"
	beginDate, _ := strconv.ParseInt(beginDateStr, 10, 64)
	return beginDate
}

// CurrentDayOfMonth current time
func CurrentDayOfMonth(tm time.Time) int64 {
	formattedTime := tm.Format("20060102")
	date, _ := strconv.ParseInt(formattedTime, 10, 64)
	return date
}

// AllDaysOfMonth days of the month
func AllDaysOfMonth(date time.Time) int {
	year, month, _ := date.Date()
	days := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	return days
}

// DayOfMonth day of month
func DayOfMonth(date time.Time) int {
	dayOfMonth := date.Day()
	return dayOfMonth
}

// LastDayOfMonth the last day of the month
func LastDayOfMonth(date time.Time) int {
	year, month, _ := date.Date()
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
	return lastDay.Day()
}

func TimeLineFormat(tm time.Time) string {
	return tm.Format(timeLineFormat)

}
