package uid

// MakeExtID makes extentID.
func MakeExtID(groupID, groupSeq uint16) uint32 {
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
