package log

import (
  "github.com/hashicorp/raft"
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
