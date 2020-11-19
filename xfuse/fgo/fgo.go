// Package fgo wraps github.com/hanwen/go-xfuse satisfying xfuse.FS.
//
// It's the default go xfuse implementation in present.
package fgo

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func shit() {
	fs.Options{
		MountOptions:      fuse.MountOptions{},
		EntryTimeout:      nil,
		AttrTimeout:       nil,
		NegativeTimeout:   nil,
		FirstAutomaticIno: 0,
		OnAdd:             nil,
		NullPermissions:   false,
		UID:               0,
		GID:               0,
		ServerCallbacks:   nil,
		Logger:            nil,
	}
}
