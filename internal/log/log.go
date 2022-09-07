package log

import (
  "fmt"
  api "github.com/oku3san/proglog/api/v1"
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

func (l *Log) Append(record *api.Record) (uint64, error) {
  l.mu.Lock()
  defer l.mu.Unlock()

  highestOffset, err := l.highestOffset()
  if err != nil {
    return 0, err
  }

  if l.activeSegment.IsMaxed() {
    err = l.newSegnemt(highestOffset + 1)
    if err != nil {
      return 0, err
    }
  }

  off, err := l.activeSegment.Append(record)
  if err != nil {
    return 0, err
  }

  return off, err
}

func (l *Log) Read(off uint64) (*api.Record, error) {
  l.mu.RLock()
  defer l.mu.RUnlock()
  var s *segment
  for _, segment := range l.segments {
    if segment.baseOffset <= off && off < segment.nextOffset {
      s = segment
      break
    }
  }
  if s == nil || s.nextOffset <= off {
    return nil, fmt.Errorf("offset out of range: %d", off)
  }
  return s.Read(off)
}

func (l *Log) Close() error {
  l.mu.Lock()
  defer l.mu.Unlock()
  for _, segment := range l.segments {
    if err := segment.Close(); err != nil {
      return err
    }
  }
  return nil
}

func (l *Log) LowestOffset() (uint64, error) {
  l.mu.RLock()
  defer l.mu.RUnlock()
  return l.segments[0].baseOffset, nil
}

func (l *Log) HighestOffset() (uint64, error) {
  l.mu.RLock()
  defer l.mu.RUnlock()
  return l.HighestOffset()
}

func (l *Log) Truncate(lowest uint64) error {
  l.mu.Lock()
  defer l.mu.Unlock()
  var segments []*segment
  for _, s := range l.segments {
    if s.nextOffset <= lowest+1 {
      if err := s.Remove(); err != nil {
        return err
      }
      continue
    }
    segments = append(segments, s)
  }
  l.segments = segments
  return nil
}
