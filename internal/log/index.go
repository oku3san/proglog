package log

import (
  "github.com/tysonmote/gommap"
  "os"
)

const (
  offWidth uint64 = 4
  posWidth uint64 = 8
  entWidth        = offWidth + posWidth
)

type index struct {
  file *os.File
  mmap gommap.MMap
  size uint64
}

func newIndex(f *os.File, c Config) (*index, error) {
  idx := &index{
    file: f,
  }
  fi, err := os.Stat(f.Name())
  if err != nil {
    return nil, err
  }
  idx.size = uint64(fi.Size())
  if err = os.Truncate(
    f.Name(), int64(c.Segment.MaxIndexBytes),
  ); err != nil {
    return nil, err
  }
  if idx.mmap, err = gommap.Map(
    idx.file.Fd(),
    gommap.PROT_READ|gommap.PROT_WRITE,
    gommap.MAP_SHARED,
  ); err != nil {
    return nil, err
  }
  return idx, nil
}
