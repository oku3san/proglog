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

func (i *index) Close() error {
  if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
    return err
  }
  if err := i.file.Sync(); err != nil {
    return err
  }
  if err := i.file.Truncate(int64(i.size)); err != nil {
    return err
  }
  return i.file.Close()
}
