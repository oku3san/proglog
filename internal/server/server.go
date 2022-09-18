package server

import api "github.com/oku3san/proglog/api/v1"

type Config struct {
  CommitLog CommitLog
}

var _ api.LogServer = (*grpcServer)(nil)

type grpcServer struct {
  api.UnimplementedLogServer
  *Config
}

func newgrpcServer(config *Config) (sev *grpcServer, err error) {
  sev = &grpcServer{
    Config: config,
  }
  return sev, nil
}
