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
	"testing"
	"time"
)

func Test_Hour(t *testing.T) {
	t.Logf("hour: %v", Hour(time.Now()))
}

func Test_Yesterday(t *testing.T) {
	t.Logf("hour: %v", YesterdayHour())
}

func Test_BeforeYesterdayHour(t *testing.T) {
	t.Logf("hour: %v", BeforeYesterdayHour())
}

func Test_TimeHour(t *testing.T) {
	now := Hour(time.Now())
	yesterday := YesterdayHour()
	beforeYesterday := BeforeYesterdayHour()
	t.Logf("now: %v", TimeHourInt(now))
	t.Logf("yesterday: %v", TimeHourInt(yesterday))
	t.Logf("beforeYesterday: %v", TimeHourInt(beforeYesterday))
}

func Test_TimeLineFormat(t *testing.T) {

	now := time.Now()
	formattedTime := now.Format("2006-01-02 15:04:05")
	t.Logf("time %v", TimeLineFormat(time.Now()))
	t.Logf("time %v", formattedTime)

}
