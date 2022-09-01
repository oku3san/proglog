package log

import "sync"

type Log struct {
  mu            sync.RWMutex
  Dir           string
  Config        Config
  activeSegment *segment
  segments      []*segment
}

func NewLog(dir string, c Config) (*Log, error) {
  if c.Segment.MaxStoreBytes == 0 {
    c.Segment.MaxStoreBytes = 1024
  }
  if c.Segment.MaxIndexBytes == 0 {
    c.Segment.MaxIndexBytes = 2014
  }
  l := &Log{
    Dir:    dir,
    Config: c,
  }
  return l, l.setup()
}
