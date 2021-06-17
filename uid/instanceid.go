package uid

import (
	"fmt"
	"regexp"
	"strings"

	"g.tesamc.com/IT/zaipkg/xmath/xrand"
)

const InstanceIDLen = 14

// GetIDCFromInstanceID gets idc label from instance_id.
// <region>-<city>-<idc_number>-<machine_number>
// e.g. cn-sz-001-0001, the return value will be cn-sz-001
func GetIDCFromInstanceID(instanceID string) string {

	ss := strings.Split(instanceID, "-")
	if len(ss) != 4 {
		panic(fmt.Sprintf("illegal instance_id: %s", instanceID))
	}

	return strings.TrimSuffix(instanceID, "-"+ss[3])
}

// GetMachineFromInstanceID gets machine_number from instance_id.
// <region>-<city>-<idc_number>-<machine_number>
// e.g. cn-sz-001-0001, the return value will be 0001
func GetMachineFromInstanceID(instanceID string) string {
	ss := strings.Split(instanceID, "-")
	if len(ss) != 4 {
		panic(fmt.Sprintf("illegal instance_id: %s", instanceID))
	}

	return ss[3]
}

// GenRandInstanceID generates an instance_id for testing only.
func GenRandInstanceID() string {

	return MakeInstanceID("cn", "sz", xrand.Int63n(1000), xrand.Int63n(1000))
}

// GenSeqInstanceID generates sequential machine number(from 1) with default idc_code instance_ids.
func GenSeqInstanceID(cnt int) []string {

	ret := make([]string, cnt)

	for i := 1; i <= cnt; i++ {
		ret[i-1] = MakeInstanceID("cn", "sz", 1, int64(i))
	}
	return ret
}

func MakeInstanceID(region, city string, idc, machine int64) string {

	if len(region) != 2 {
		panic("illegal region")
	}

	if len(city) != 2 {
		panic("illegal city")
	}

	if idc > 999 {
		panic("illegal idc")
	}

	if machine > 9999 {
		panic("illegal machine")
	}

	return fmt.Sprintf("%s-%s-%03d-%04d", region, city, idc, machine)
}

var InstanceIDRegexp = regexp.MustCompile(`^[a-z]{2}-[a-z]{2}-\d{3}-\d{4}`)

// IsValidInstanceID returns the instanceID is valid in Zai or not.
func IsValidInstanceID(instanceID string) bool {

	if len(instanceID) != InstanceIDLen {
		return false
	}

	return InstanceIDRegexp.FindString(instanceID) == instanceID
}
