package server

import (
  "context"
  api "github.com/oku3san/proglog/api/v1"
)

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

func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (
  *api.ProduceResponse, error) {
  offset, err := s.CommitLog.Append(req.Record)
  if err != nil {
    return nil, err
  }
  return &api.ProduceResponse{Offset: offset}, nil
}

func (s *grpcServer) Consume(ctx context.Context, req *api.ConsumeRequest) (
  *api.ConsumeResponse, error) {
  record, err := s.CommitLog.Read(req.Offset)
  if err != nil {
    return nil, err
  }
  return &api.ConsumeResponse{Record: record}, nil
}

func (s *grpcServer) ProduceStream(
  stream api.Log_ProduceStreamServer,
) error {
  for {
    req, err := stream.Recv()
    if err != nil {
      return err
    }
    res, err := s.Produce(stream.Context(), req)
    if err != nil {
      return err
    }
    if err = stream.Send(res); err != nil {
      return err
    }
  }
}
