package main

import (
  "hash"
  "hash/adler32"
  "hash/crc32"
  "crypto/md5"
  "crypto/sha1"
  "crypto/sha256"
  "fmt"
  "io"
)

type MultiHashContext struct {
  adler32 hash.Hash
  crc32 hash.Hash
  md5 hash.Hash
  sha1 hash.Hash
  sha256 hash.Hash
}

func New() MultiHashContext {
  return MultiHashContext{
    adler32: adler32.New(),
    crc32: crc32.NewIEEE(),
    md5: md5.New(),
    sha1: sha1.New(),
    sha256: sha256.New() }
}

func (h *MultiHashContext) Writer() io.Writer {
  return io.MultiWriter(
    h.adler32,
    h.crc32,
    h.md5,
    h.sha1,
    h.sha256)
}

func (h *MultiHashContext) Result() map[string]string {
  result := make(map[string]string)
  result["adler32"] = fmt.Sprintf("%x", h.adler32.Sum(nil))
  result["crc32"] = fmt.Sprintf("%x", h.crc32.Sum(nil))
  result["md5"] = fmt.Sprintf("%x", h.md5.Sum(nil))
  result["sha1"] = fmt.Sprintf("%x", h.sha1.Sum(nil))
  result["sha256"] = fmt.Sprintf("%x", h.sha256.Sum(nil))
  return result
}

func main() {
  h := New()
  w := h.Writer()
  fmt.Fprintf(w, "Hello, world!")
  fmt.Printf("%v", h.Result())
}
