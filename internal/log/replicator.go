package log

import (
  api "github.com/oku3san/proglog/api/v1"
  "go.uber.org/zap"
  "google.golang.org/grpc"
  "sync"
)

type Replicator struct {
  DialOptions []grpc.DialOption
  LocalServer api.LogClient

  logger *zap.Logger

  mu      sync.Mutex
  servers map[string]chan struct{}
  closed  bool
  close   chan struct{}
}
