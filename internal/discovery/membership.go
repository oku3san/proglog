package discovery

import (
  "github.com/hashicorp/serf/serf"
  "go.uber.org/zap"
)

type Membership struct {
  Config
  handler Handler
  serf    *serf.Serf
  events  chan serf.Event
  logger  *zap.Logger
}

func New(handler Handler, config Config) (*Membership, error) {
  c := &Membership{
    Config:  config,
    handler: handler,
    logger:  zap.L().Named("membership"),
  }
  if err := c.setupSerf(); err != nil {
    return nil, err
  }
  return c, nil
}

type Config struct {
  NodeName       string
  BindAddr       string
  Tags           map[string]string
  StartJoinAddrs []string
}
