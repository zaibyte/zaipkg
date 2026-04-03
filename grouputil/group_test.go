package grouputil

import (
	"sort"
	"testing"

	"github.com/zaibyte/zproto/pkg/metapb"
)

func TestGroupsSort(t *testing.T) {
	gs := make([]*metapb.Group, 10)

	for i := range gs {
		gs[i] = &metapb.Group{Avail: uint64(i)}
	}
	sort.Sort(GroupsAvail(gs))
	for i := range gs {
		if gs[i].Avail != 9-uint64(i) {
			t.Fatal("sort is not desc")
		}
	}
}
