package log

import (
  api "github.com/oku3san/proglog/api/v1"
  "github.com/stretchr/testify/require"
  "os"
  "testing"
)

func TestLog(t *testing.T) {
  for scenario, fn := range map[string]func(
    t *testing.T, log *Log,
  ){
    "append and read a record succeeds": testAppendRead,
    "offset out of range error":         testOutOfRangeErr,
    "init with existing segments":       testInitExisting,
    "reader":                            testReader,
    "truncate":                          testTruncate,
  } {
    t.Run(scenario, func(t *testing.T) {
      dir, err := os.MkdirTemp("", "store-test")
      require.NoError(t, err)
      defer os.RemoveAll(dir)

      c := Config{}
      c.Segment.MaxStoreBytes = 32
      log, err := NewLog(dir, c)
      require.NoError(t, err)

      fn(t, log)
    })
  }
}

func testAppendRead(t *testing.T, log *Log) {
  append := &api.Record{
    Value: []byte("hello world"),
  }
  off, err := log.Append(append)
  require.NoError(t, err)
  require.Equal(t, uint64(0), off)

  read, err := log.Read(off)
  require.NoError(t, err)
  require.Equal(t, append.Value, read.Value)
  require.NoError(t, log.Close())
}
