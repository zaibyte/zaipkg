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
