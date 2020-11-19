package xfuse

// Config is the fuse configs.
type Config struct {
}

// MountOptions is the fuse mount options.
type MountOptions struct {
	AllowOther bool

	// Options are passed as -o string to fusermount.
	Options []string

	// Default is _DEFAULT_BACKGROUND_TASKS, 12.  This numbers
	// controls the allowed number of requests that relate to
	// async I/O.  Concurrency for synchronous I/O is not limited.
	MaxBackground int

	// Write size to use.  If 0, use default. This number is
	// capped at the kernel maximum.
	MaxWrite int

	// Max read ahead to use.  If 0, use default. This number is
	// capped at the kernel maximum.
	MaxReadAhead int

	// If IgnoreSecurityLabels is set, all security related xattr
	// requests will return NO_DATA without passing through the
	// user defined filesystem.  You should only set this if you
	// file system implements extended attributes, and you are not
	// interested in security labels.
	IgnoreSecurityLabels bool // ignoring labels should be provided as a fusermount mount option.

	// If RememberInodes is set, we will never forget inodes.
	// This may be useful for NFS.
	RememberInodes bool

	// Values shown in "df -T" and friends
	// First column, "Filesystem"
	FsName string

	// Second column, "Type", will be shown as "fuse." + Name
	Name string

	// If set, wrap the file system in a single-threaded locking wrapper.
	SingleThreaded bool

	// If set, return ENOSYS for Getxattr calls, so the kernel does not issue any
	// Xattr operations at all.
	DisableXAttrs bool

	// If set, print debugging information.
	Debug bool

	// If set, ask kernel to forward file locks to FUSE. If using,
	// you must implement the GetLk/SetLk/SetLkw methods.
	EnableLocks bool

	// If set, ask kernel not to do automatic data cache invalidation.
	// The filesystem is fully responsible for invalidating data cache.
	ExplicitDataCacheControl bool

	// If set, fuse will first attempt to use syscall.Mount instead of
	// fusermount to mount the filesystem. This will not update /etc/mtab
	// but might be needed if fusermount is not available.
	DirectMount bool

	// Options passed to syscall.Mount, the default value used by fusermount
	// is syscall.MS_NOSUID|syscall.MS_NODEV
	DirectMountFlags uintptr
}
