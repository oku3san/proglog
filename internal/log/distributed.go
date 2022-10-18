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

}
