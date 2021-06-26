package uid

// MakeExtID makes extentID.
// groupSeq must >= 1.
func MakeExtID(groupID, groupSeq uint16) uint32 {
	if groupSeq < 1 {
		panic("group_seq in ext_id must >= 1")
	}
	return uint32(groupSeq)<<16 | uint32(groupID)
}

const MaxGroupSeq = (1 << 16) - 1

// ParseExtID gets groupID and its seq in this group.
func ParseExtID(extID uint32) (groupID, groupSeq uint16) {
	groupID = uint16(extID) & MaxGroupID
	groupSeq = uint16((extID >> 16) & MaxGroupSeq)
	return
}

// GetGroupID gets groupID from extID.
func GetGroupID(extID uint32) uint16 {
	return uint16(extID) & MaxGroupID
}
