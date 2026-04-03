package xraft

import "github.com/zaibyte/zaipkg/config"

// Config of Raft.
type Config struct {
	RaftNode RaftNodeConfig `toml:"raft_node"`
	Member   MemberConfig   `toml:"member"`
	NodeHost NodeHostConfig `toml:"node_host"`
}

func (cfg *Config) Adjust() {
	cfg.RaftNode.adjust()
	cfg.NodeHost.adjust()
}

const (
	defaultElectionRTT        = 10
	defaultHeartbeatRTT       = 2
	defaultCompactionOverhead = 1024
	defaultMaxInMemLogSize    = 128 * 1024 * 1024
	defaultRTTMillisecond     = 2 // 2ms is a good start, the ping between two server is about 300-400us in anlian.
)

type RaftNodeConfig struct {
	// NodeID is a non-zero value used to identify a node within a Raft cluster.
	NodeID uint64 `toml:"node_id"`
	// ClusterID is the unique value used to identify a Raft cluster.
	ClusterID uint64 `toml:"cluster_id"`
	// CheckQuorum specifies whether the leader node should periodically check
	// non-leader node status and step down to become a follower node when it no
	// longer has the quorum.
	CheckQuorum bool `toml:"check_quorum"`
	// ElectionRTT is the minimum number of message RTT between elections. Message
	// RTT is defined by NodeHostConfig.RTTMillisecond. The Raft paper suggests it
	// to be a magnitude greater than HeartbeatRTT, which is the interval between
	// two heartbeats. In Raft, the actual interval between elections is
	// randomized to be between ElectionRTT and 2 * ElectionRTT.
	//
	// As an example, assuming NodeHostConfig.RTTMillisecond is 100 millisecond,
	// to set the election interval to be 1 second, then ElectionRTT should be set
	// to 10.
	//
	// When CheckQuorum is enabled, ElectionRTT also defines the interval for
	// checking leader quorum.
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
	// automcatically. It is defined in terms of the number of applied Raft log
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
	CompactionOverhead uint64 `toml:"compaction_overhead"`
	// MaxInMemLogSize is the target size in bytes allowed for storing in memory
	// Raft logs on each Raft node. In memory Raft logs are the ones that have
	// not been applied yet.
	// MaxInMemLogSize is a target value implemented to prevent unbounded memory
	// growth, it is not for precisely limiting the exact memory usage.
	// When MaxInMemLogSize is 0, the target is set to math.MaxUint64. When
	// MaxInMemLogSize is set and the target is reached, error will be returned
	// when clients try to make new proposals.
	// MaxInMemLogSize is recommended to be significantly larger than the biggest
	// proposal you are going to use.
	MaxInMemLogSize uint64 `toml:"max_in_mem_log_size"`
	// DisableAutoCompactions disables auto compaction used for reclaiming Raft
	// entry storage spaces. By default, compaction is issued every time when
	// a snapshot is captured, this helps to reclaim disk spaces as soon as
	// possible at the cost of higher IO overhead. Users can disable such auto
	// compactions and use NodeHost.RequestCompaction to manually request such
	// compactions when necessary.
	DisableAutoCompactions bool `toml:"disable_auto_compactions"`
}

func (cfg *RaftNodeConfig) adjust() {
	config.Adjust(&cfg.CompactionOverhead, defaultCompactionOverhead)
	config.Adjust(&cfg.ElectionRTT, defaultElectionRTT)
	config.Adjust(&cfg.HeartbeatRTT, defaultHeartbeatRTT)
	config.Adjust(&cfg.MaxInMemLogSize, defaultMaxInMemLogSize)
}

type MemberConfig struct {
	//  - starting a brand new Raft cluster, set join to false and specify all initial
	//    member node details in the initialMembers map.
	//  - joining a new node to an existing Raft cluster, set join to true and leave
	//    the initialMembers map empty. This requires the joining node to have already
	//    been added as a member node of the Raft cluster.
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

func (cfg *NodeHostConfig) adjust() {
	config.Adjust(cfg.RTTMillisecond, defaultRTTMillisecond)
}
