package uid

import (
	"fmt"
	"regexp"
	"strings"

	"g.tesamc.com/IT/zaipkg/xmath/xrand"
)

// GetIDCFromInstanceID gets idc label from instance_id.
// <region>-<city>-<idc_number>-<machine_number>
// e.g. cn-sz-001-0000001
func GetIDCFromInstanceID(instanceID string) string {

	ss := strings.Split(instanceID, "-")
	if len(ss) != 4 {
		panic(fmt.Sprintf("illegal instance_id: %s", instanceID))
	}

	return strings.TrimSuffix(instanceID, "-"+ss[3])
}

// GenRandInstanceID generates an instance_id for testing only.
func GenRandInstanceID() string {
	return fmt.Sprintf("cn-sz-001-%02d", xrand.Int63n(9999999))
}

var InstanceIDRegexp = regexp.MustCompile(`^[a-z]{2}-[a-z]{2}-\d{3}-\d{7}`)

// IsValidInstanceID returns the instanceID is valid in Zai or not.
func IsValidInstanceID(instanceID string) bool {
	return InstanceIDRegexp.FindString(instanceID) == instanceID
}
