package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"hash/adler32"
	"hash/crc32"
	"io"
)

type MultiHashContext struct {
	contexts map[string]hash.Hash
}

func New() MultiHashContext {
	contexts := make(map[string]hash.Hash)
	contexts["adler32"] = adler32.New()
	contexts["crc32"] = crc32.NewIEEE()
	contexts["md5"] = md5.New()
	contexts["sha1"] = sha1.New()
	contexts["sha256"] = sha256.New()
	return MultiHashContext{contexts: contexts}
}

func (h *MultiHashContext) Writer() io.Writer {
	var elements []io.Writer
	for _, v := range h.contexts {
		elements = append(elements, v)
	}
	return io.MultiWriter(elements...)
}

func (h *MultiHashContext) Result() map[string]string {
	result := make(map[string]string)
	for k, v := range h.contexts {
		result[k] = fmt.Sprintf("%x", v.Sum(nil))
	}
	return result
}

func main() {
	h := New()
	w := h.Writer()
	fmt.Fprintf(w, "Hello, world!")
	fmt.Printf("%v", h.Result())
}
