package grouputil

import (
	"g.tesamc.com/IT/zproto/pkg/metapb"
)

// GroupsAvail implements sort.Interface for sorting avail of groups.
type GroupsAvail []*metapb.Group

func (g GroupsAvail) Len() int {

	return len(g)
}

func (g GroupsAvail) Less(i, j int) bool {
	if g[i].Avail > g[j].Avail {
		return true
	}
	return false
}

func (g GroupsAvail) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

// SetStateByExt sets group state by its extents.
// Call it when extents changed.
func SetStateByExt(g *metapb.Group, replicas int) {

	defer func() {
		if g.GetState() == metapb.GroupState_Group_Collapse {
			for _, ext := range g.Exts {
				delete(g.Exts, ext.GetId())
			}
		}
	}()

	extCnt := len(g.GetExts())

	if extCnt == 0 { // Broken extent has been removed from group.
		g.State = metapb.GroupState_Group_Collapse
		return
	}

	rwCnt, fullCnt, sealCnt := 0, 0, 0
	for _, ext := range g.Exts {
		es := ext.GetState()
		switch es {
		case metapb.ExtentState_Extent_ReadWrite:
			rwCnt++
		case metapb.ExtentState_Extent_Full:
			fullCnt++
		case metapb.ExtentState_Extent_Sealed:
			sealCnt++
		}
	}

	if rwCnt+fullCnt+sealCnt == 0 {
		g.State = metapb.GroupState_Group_Collapse
		return
	}

	if rwCnt >= replicas {
		g.State = metapb.GroupState_Group_ReadWrite
		return
	}

	g.State = metapb.GroupState_Group_Read
	return
}
