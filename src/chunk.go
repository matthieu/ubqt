package ubqt

import (
  "gob" // until custom serialization gets implemented
  "os"
)

// In-memory representation of a chunk. Not necessarily ready for execution just
// yet but should be expanded enough to make all introspective operations easy.

type Chunk struct {
  header  Header
  topfn   Funk
}

type Header struct {
  sig     string
  version byte
  format  byte
}

type Funk struct {
  sourceName  string    // source file for top level funks
  firstLine   int32
  lastLine    int32
  upvalNum    byte
  paramNum    byte
  varargFlag  byte
  maxStack    byte
  code        []int32
  constants   []*Value
  funks       []*Funk
  srcLines    []int32   // source line number for each instruction, optional
  locals      []Local   // may be empty (optional)
  upvalues    []string  // may be empty (optional)
}

type Local struct {
  name  string
  start int32
  end   int32
}

func (chunk *Chunk) serializeFile(fileName string) {
  f, err := os.Open(fileName, os.O_WRONLY, 664)
  if err != nil {
    panic("Error when opening file on write " + fileName + ": " + err.String())
  }
  enc := gob.NewEncoder(f)
  err = enc.Encode(chunk)
  f.Close()
  if err != nil {
    panic("Error when encoding chunk to " + fileName + ": " + err.String())
  }
  return
}

func deserializeFile(fileName string) (chunk *Chunk) {
  f, err := os.Open(fileName, os.O_RDONLY, 0)
  if err != nil {
    panic("Error when opening file on read " + fileName + ": " + err.String())
  }
  dec := gob.NewDecoder(f)
  err = dec.Decode(chunk)
  f.Close()
  if err != nil {
    panic("Error when decoding chunk from " + fileName + ": " + err.String())
  }
  return
}
