package extutil

import "g.tesamc.com/IT/zaipkg/config/settings"

const (
	kb = 1024
	mb = 1024 * kb
	gb = 1024 * mb
)

// ExtPreallocate is the disk size maybe taken by an extent.
// Used in Keeper when it want to create a new extent on a certain disk,
// after picking up disk, we should updating the disk usage by this size,
// then beginning to pick up the next disk.
// It's not a precise number, but it's okay because the more accurate usage will be report by disk heartbeat.
var ExtPreallocate = map[uint16]uint64{
	settings.ExtV1: (256 + 1) * gb,
}
