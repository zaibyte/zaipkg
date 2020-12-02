package xraft

// Raft config.
// TODO set CheckQuorum = true
type Config struct {
	// NodeID is a non-zero value used to identify a node within a Raft cluster.
	NodeID uint64 `toml:"node_id"`
	// ClusterID is the unique value used to identify a Raft cluster.
	ClusterID   uint64 `toml:"cluster_id"`
	ElectionRTT uint64 `toml:"election_rtt"`
	// HeartbeatRTT is the number of message RTT between heartbeats. Message
	// RTT is defined by NodeHostConfig.RTTMillisecond. The Raft paper suggest the
	// heartbeat interval to be close to the average RTT between nodes.
	//
	// As an example, assuming NodeHostConfig.RTTMillisecond is 100 millisecond,
	// to set the heartbeat interval to be every 200 milliseconds, then
	// HeartbeatRTT should be set to 2.
	HeartbeatRTT uint64 `toml:"heartbeat_rtt"`
	// SnapshotEntries defines how often the state machine should be snapshotted
	// automatically. It is defined in terms of the number of applied Raft log
	// entries. SnapshotEntries can be set to 0 to disable such automatic
	// snapshotting.
	//
	// When SnapshotEntries is set to N, it means a snapshot is created for
	// roughly every N applied Raft log entries (proposals). This also implies
	// that sending N log entries to a follower is more expensive than sending a
	// snapshot.
	//
	// Once a snapshot is generated, Raft log entries covered by the new snapshot
	// can be compacted. This involves two steps, redundant log entries are first
	// marked as deleted, then they are physically removed from the underlying
	// storage when a LogDB compaction is issued at a later stage. See the godoc
	// on CompactionOverhead for details on what log entries are actually removed
	// and compacted after generating a snapshot.
	//
	// Once automatic snapshotting is disabled by setting the SnapshotEntries
	// field to 0, users can still use NodeHost's RequestSnapshot or
	// SyncRequestSnapshot methods to manually request snapshots.
	SnapshotEntries uint64 `toml:"snapshot_entries"`
	// CompactionOverhead defines the number of most recent entries to keep after
	// each Raft log compaction. Raft log compaction is performance automatically
	// every time when a snapshot is created.
	//
	// For example, when a snapshot is created at let's say index 10,000, then all
	// Raft log entries with index <= 10,000 can be removed from that node as they
	// have already been covered by the created snapshot image. This frees up the
	// maximum storage space but comes at the cost that the full snapshot will
	// have to be sent to the follower if the follower requires any Raft log entry
	// at index <= 10,000. When CompactionOverhead is set to say 500, Dragonboat
	// then compacts the Raft log up to index 9,500 and keeps Raft log entries
	// between index (9,500, 1,0000]. As a result, the node can still replicate
	// Raft log entries between index (9,500, 1,0000] to other peers and only fall
	// back to stream the full snapshot if any Raft log entry with index <= 9,500
	// is required to be replicated.
	CompactionOverhead uint64

	Member   MemberConfig   `toml:"member"`
	NodeHost NodeHostConfig `toml:"node_host"`
}

//  - starting a brand new Raft cluster, set join to false and specify all initial
//    member node details in the initialMembers map.
//  - joining a new node to an existing Raft cluster, set join to true and leave
//    the initialMembers map empty. This requires the joining node to have already
//    been added as a member node of the Raft cluster.
type MemberConfig struct {
	InitialMembers map[uint64]string `toml:"initial_members"`
	Join           bool              `toml:"join"`
}

type NodeHostConfig struct {
	// WALDir is the directory used for storing the WAL of Raft entries. It is
	// recommended to use low latency storage such as NVME SSD with power loss
	// protection to store such WAL data. Leave WALDir to have zero value will
	// have everything stored in NodeHostDir.
	WALDir string `toml:"wal_dir"`
	// NodeHostDir is where everything else is stored.
	NodeHostDir string `toml:"node_host_dir"`
	// RTTMillisecond defines the average Round Trip Time (RTT) in milliseconds
	// between two NodeHost instances. Such a RTT interval is internally used as
	// a logical clock tick, Raft heartbeat and election intervals are both
	// defined in term of how many such RTT intervals.
	// Note that RTTMillisecond is the combined delays between two NodeHost
	// instances including all delays caused by network transmission, delays
	// caused by NodeHost queuing and processing. As an example, when fully
	// loaded, the average Round Trip Time between two of our NodeHost instances
	// used for benchmarking purposes is up to 500 microseconds when the ping time
	// between them is 100 microseconds. Set RTTMillisecond to 1 when it is less
	// than 1 ms in your environment.
	RTTMillisecond uint64 `toml:"rtt_millisecond"`
	// RaftAddress is a hostname:port or IP:port address used by the Raft RPC
	// module for exchanging Raft messages and snapshots. This is also the
	// identifier for a NodeHost instance. RaftAddress should be set to the
	// public address that can be accessed from remote NodeHost instances.
	RaftAddress string `toml:"raft_address"`
}
