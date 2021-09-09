package uid

// MakeExtID makes extentID.
// groupSeq must >= 1.
func MakeExtID(groupID uint32, groupSeq uint16) uint32 {
	if groupSeq < 1 {
		panic("group_seq in ext_id must >= 1")
	}
	return uint32(groupSeq)<<19 | groupID
}

const MaxGroupSeq = (1 << 13) - 1

// ParseExtID gets groupID and its seq in this group.
func ParseExtID(extID uint32) (groupID uint32, groupSeq uint16) {
	groupID = extID & MaxGroupID
	groupSeq = uint16((extID >> 19) & MaxGroupSeq)
	return
}

// GetGroupID gets groupID from extID.
func GetGroupID(extID uint32) uint32 {
	return extID & MaxGroupID
}
