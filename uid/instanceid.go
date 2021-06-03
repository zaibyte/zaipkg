package uid

import (
	"fmt"
	"strings"
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
