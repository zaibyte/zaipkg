package orpc

import (
	"fmt"
	"testing"
	"time"

	"g.tesamc.com/IT/zaipkg/xtest"
)

// Reference:
// 0KB:
// 100ms
// 174.390106ms
// 344.181861ms
// 446.971682ms
// 1.46891831s
// 3s
//
// 1KB:
// 100ms
// 197.171423ms
// 435.151004ms
// 701.576989ms
// 1.140340844s
// 3s
//
// 128KB:
// 100ms
// 221.41572ms
// 478.453162ms
// 590.148055ms
// 1.175595984s
// 3s
//
// 512KB:
// 100ms
// 223.487535ms
// 481.533594ms
// 1.023797481s
// 2.236380328s
// 3s
//
// 4MB:
// 100ms
// 684.928948ms
// 1.206078366s
// 2.117665492s
// 3s
// 3s
//
// 16MB:
// 100ms
// 781.0251ms
// 1.697363245s
// 2.918559311s
// 3s
// 3s
func TestRetryer_GetSleepDurationReasonable(t *testing.T) {

	if !xtest.IsPropEnabled() {
		t.Skip("skip this test, because have passed the prop testing")
	}

	r := &Retryer{
		MinSleep: 100 * time.Millisecond, // For most Zai request, the requests will be light, so the sleep shouldn't be high.
		MaxTried: 5,
		MaxSleep: 3 * time.Second,
	}

	for i := 0; i < 6; i++ {
		fmt.Println(r.GetSleepDuration(i, 0))
	}
	for i := 0; i < 6; i++ {
		fmt.Println(r.GetSleepDuration(i, 1024))
	}
	for i := 0; i < 6; i++ {
		fmt.Println(r.GetSleepDuration(i, 128*1024))
	}
	for i := 0; i < 6; i++ {
		fmt.Println(r.GetSleepDuration(i, 512*1024))
	}
	for i := 0; i < 6; i++ {
		fmt.Println(r.GetSleepDuration(i, 4*1024*1024))
	}
	for i := 0; i < 6; i++ {
		fmt.Println(r.GetSleepDuration(i, 16*1024*1024))
	}
}
