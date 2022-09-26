package server

import (
  "context"
  api "github.com/oku3san/proglog/api/v1"
  "github.com/oku3san/proglog/internal/log"
  "github.com/stretchr/testify/require"
  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials/insecure"
  "google.golang.org/grpc/status"
  "net"
  "os"
  "testing"
)

func TestServer(t *testing.T) {
  for scenario, fn := range map[string]func(
    t *testing.T,
    client api.LogClient,
    config *Config,
  ){
    "produce.consume a message to/from the log succeeds": testProduceConsume,
    "produce/consume stream succeeds":                    testProduceConsumeStream,
    "consume past log boundary fails":                    testConsumePastBoundary,
  } {
    t.Run(scenario, func(t *testing.T) {
      client, config, teardown := setupTest(t, nil)
      defer teardown()
      fn(t, client, config)
    })
  }
}

func setupTest(t *testing.T, fn func(config *Config)) (
  client api.LogClient,
  cfg *Config,
  teardown func(),
) {
  t.Helper()

  l, err := net.Listen("tcp", ":0")
  require.NoError(t, err)

  clientOptions := []grpc.DialOption{
    grpc.WithTransportCredentials(insecure.NewCredentials())}
  cc, err := grpc.Dial(l.Addr().String(), clientOptions...)
  require.NoError(t, err)

  dir, err := os.MkdirTemp("", "server-test")
  require.NoError(t, err)

  clog, err := log.NewLog(dir, log.Config{})
  require.NoError(t, err)

  cfg = &Config{
    CommitLog: clog,
  }
  if fn != nil {
    fn(cfg)
  }
  server, err := NewGRPCServer(cfg)
  require.NoError(t, err)

  go func() {
    server.Serve(l)
  }()

  client = api.NewLogClient(cc)

  return client, cfg, func() {
    cc.Close()
    server.Stop()
    l.Close()
    clog.Remove()
  }
}

func testProduceConsume(t *testing.T, client api.LogClient, config *Config) {
  ctx := context.Background()

  want := &api.Record{
    Value: []byte("hello world"),
  }

  produce, err := client.Produce(
    ctx,
    &api.ProduceRequest{
      Record: want,
    },
  )
  require.NoError(t, err)
  want.Offset = produce.Offset

  consume, err := client.Consume(ctx, &api.ConsumeRequest{
    Offset: produce.Offset,
  })
  require.NoError(t, err)
  require.Equal(t, want.Value, consume.Record.Value)
  require.Equal(t, want.Offset, consume.Record.Offset)
}

func testConsumePastBoundary(
  t *testing.T,
  cliet api.LogClient,
  config *Config,
) {
  ctx := context.Background()

  produce, err := cliet.Produce(ctx, &api.ProduceRequest{
    Record: &api.Record{
      Value: []byte("hello world"),
    },
  })
  require.NoError(t, err)

  consume, err := cliet.Consume(ctx, &api.ConsumeRequest{
    Offset: produce.Offset + 1,
  })
  if consume != nil {
    t.Fatal("consume not nil")
  }
  got := status.Code(err)
  want := status.Code(api.ErrOffsetOutOfRange{}.GRPCStatus().Err())
  if got != want {
    t.Fatalf("got err: %v, want: %v", got, want)
  }
}
