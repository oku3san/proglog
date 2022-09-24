package server

import (
  api "github.com/oku3san/proglog/api/v1"
  "testing"
)

func TestServer(t *testing.T) {
  for scenario, fn := range map[string]func(
    t *testing.T,
    client api.LogClient,
    config *Config,
  ){
    "produce.consume a message to/from the log succeeds": testProcudeConsume,
    "produce/consume stream succeeds": testProduceConsumeStream,
    "consume past log boundary fails": testConsumePastBoundary,
  } {
    t.Run(scenario, func(t *testimg.T) {
      client, config, teardown := setupTest(t, nil)
      defer teardown()
      fn(t, client, config)
    })
  }
}
