package log

import (
  "github.com/hashicorp/raft"
  "os"
  "path/filepath"
  "time"
)

type DistributedLog struct {
  config Config,
  log *Log,
  raftLog *logStore,
  raft *raft.Raft
}

func NewDistributedLog(dataDir string, config Config) (
  *DistributedLog,
  error,
) {
  l := &DistributedLog{
    config: config,
  }
  if err := l.setupLog(dataDir); err != nil {
    return nil, err
  }
  if err := l.setupRaft(dataDir); err != nil {
    return nil, err
  }
  return l, nil
}

func (l *DistributedLog) setupLog(dataDir string) error {
  logDir := filepath.Join(dataDir, "log")
  if err := os.MkdirAll(logDir, 0755); err != nil {
    return err
  }
  var err error
  l.log, err = NewLog(logDir, l.config)
  return err
}

func (l *DistributedLog) setupRaft(dataDir string) error {
  var err error

  fsm := &fsm{log: l.log}

  logDir := filepath.Join(dataDir, "raft", "log")
  if err := os.MkdirAll(logDir, 0755); err != nil {
    return err
  }
  logConfig := l.config
  logConfig.Segment.InitialOffset = 1
  l.raftLog, err = newLogStore(logDir, logConfig)
  if err != nil {
    return err
  }

  stableStore, err := raftboltdb.NewBoltStore(
    filepath.Join(dataDir, "raft", "stable"),
  )
  if err != nil {
    return err
  }

  retain := 1
  snapshotStore, err := raft.NewFileSnapshotStore(
    filepath.Join(dataDir, "raft"),
    retain,
    os.Stderr
  )
  if err != nil {
    return err
  }

  maxPool := 5
  timeout := 10 * time.Second
  transport := raft.NewNetworkTransport(
    l.config.Raft.StreamLayer,
    maxPool,
    timeout,
    os.Stderr,
  )
  config := raft.DefaultConfig()
  config.LocalID = l.config.Raft.LocalID
  if l.config.Raft.HeatbeatTimeout != 0 {
    config.HearbeatTimeout = l.config.Raft.HearbeatTimeout
  }
  if l.config.Raft.ElectionTimeout != 0 {
    config.ElectionTimeout = l.config.Raft.ElectioonTimeout
  }
  if l.config.Raft.LeaderLeaseTimeout != 0 {
    config.LeaderLeaseTimeout = l.config.Raft.LeaderLeaseTimeout
  }
  if l.config.Raft.CommitTimeout != 0 {
    config.CommitTimeout = l.config.Raft.CommitTimeout
  }

  l.raft, err = raft.NewRaft(
    config,
    fsm,
    l.raftLog,
    stableStore,
    snapshotStore,
    transport
  )
  if err != nil {
    return err
  }
  hasState, err := raft.HasExistingState(
    l.raftLog,
    stableStore,
    snapshotStore,
  )
  if err != nil {
    return err
  }
  if l.config.Raft.Bootstrap && !hasState {
    config := raft.Configuration{
      Servers: []raft.Server{{
        ID: config.LocalID,
        Address: transport.LocalAddr(),
      }},
    }
    err = l.raft.BootstrapCluster(config).Error()
  }
  return err

}
