package log

import (
  "os"
  "path"
  "sort"
  "strconv"
  "strings"
  "sync"
)

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

func (l *Log) setup() error {
  files, err := os.ReadDir(l.Dir)
  if err != nil {
    return err
  }
  var baseOffsets []uint64
  for _, file := range files {
    offStr := strings.TrimSuffix(
      file.Name(),
      path.Ext(file.Name()),
    )
    off, _ := strconv.ParseUint(offStr, 10, 0)
    baseOffsets = append(baseOffsets, off)
  }
  sort.Slice(baseOffsets, func(i, j int) bool {
    return baseOffsets[i] < baseOffsets[j]
  })
  for i := 0; i < len(baseOffsets); i++ {
    if err = l.newSegment(baseOffsets[i]); err != nil {
      return err
    }
    i++
  }
  if l.segments == nil {
    if err = l.newSegment(
      l.Config.Segment.InitialOffset,
    ); err != nil {
      return err
    }
    return nil
  }
}
