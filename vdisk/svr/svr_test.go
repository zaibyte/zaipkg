package svr

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/spf13/cast"

	"g.tesamc.com/IT/zaipkg/vfs"

	"github.com/stretchr/testify/assert"
)

func TestMakeDiskPath(t *testing.T) {
	root := "/root"
	diskID := uuid.NewString()
	p := MakeDiskDir(diskID, root)
	exp := filepath.Join(root, "disk_"+diskID)
	assert.Equal(t, exp, p)
}

func TestListDiskIDs(t *testing.T) {
	root, err := ioutil.TempDir(os.TempDir(), "zbuf-server")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	ids := make([]string, 1024)
	for i := range ids {
		ids[i] = uuid.NewString()
	}

	fs := vfs.GetFS()

	for _, id := range ids {
		dir := filepath.Join(root, DiskNamePrefix+cast.ToString(id))
		err = fs.MkdirAll(dir, 0777)
		if err != nil {
			t.Fatal(err)
		}
	}

	actIDs, err := ListDiskIDs(vfs.GetFS(), root)
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(ids)
	sort.Strings(actIDs)

	assert.Equal(t, ids, actIDs)
}
