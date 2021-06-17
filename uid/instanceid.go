package uid

import (
	"fmt"
	"regexp"
	"strings"

	"g.tesamc.com/IT/zaipkg/xmath/xrand"
)

const InstanceIDLen = 16

// GetIDCFromInstanceID gets idc label from instance_id.
// <region>-<city>-<idc_number>-<machine_number>
// e.g. cn-sz-001-000001, the return value will be cn-sz-001
func GetIDCFromInstanceID(instanceID string) string {

	ss := strings.Split(instanceID, "-")
	if len(ss) != 4 {
		panic(fmt.Sprintf("illegal instance_id: %s", instanceID))
	}

	return strings.TrimSuffix(instanceID, "-"+ss[3])
}

// GetMachineFromInstanceID gets machine_number from instance_id.
// <region>-<city>-<idc_number>-<machine_number>
// e.g. cn-sz-001-0000001, the return value will be 0000001
func GetMachineFromInstanceID(instanceID string) string {
	ss := strings.Split(instanceID, "-")
	if len(ss) != 4 {
		panic(fmt.Sprintf("illegal instance_id: %s", instanceID))
	}

	return ss[3]
}

// GenRandInstanceID generates an instance_id for testing only.
func GenRandInstanceID() string {
	return fmt.Sprintf("cn-sz-%03d-%06d", xrand.Int63n(1000), xrand.Int63n(1000000))
}

var InstanceIDRegexp = regexp.MustCompile(`^[a-z]{2}-[a-z]{2}-\d{3}-\d{6}`)

// IsValidInstanceID returns the instanceID is valid in Zai or not.
func IsValidInstanceID(instanceID string) bool {

	if len(instanceID) != InstanceIDLen {
		return false
	}

	return InstanceIDRegexp.FindString(instanceID) == instanceID
}
