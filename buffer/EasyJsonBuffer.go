// see https://github.com/mailru/easyjson/blob/master/buffer/pool.go
//
// I need Reset() method to test, bu private functions prevents to embed Buffer
package buffer

import (
	"sync"
)

// PoolConfig contains configuration for the allocation and reuse strategy.
type PoolConfig struct {
	StartSize  int // Minimum chunk size that is allocated.
	PooledSize int // Minimum chunk size that is reused, reusing chunks too small will result in overhead.
	MaxSize    int // Maximum chunk size that will be allocated.
}

var config = PoolConfig{
	StartSize:  128,
	PooledSize: 512,
	MaxSize:    32768,
}

// Reuse pool: chunk size -> pool.
var buffers = map[int]*sync.Pool{}

func initBuffers() {
	for l := config.PooledSize; l <= config.MaxSize; l *= 2 {
		buffers[l] = new(sync.Pool)
	}
}

func init() {
	initBuffers()
}

// Init sets up a non-default pooling and allocation strategy. Should be run before serialization is done.
func Init(cfg PoolConfig) {
	config = cfg
	initBuffers()
}

// putBuf puts a chunk to reuse pool if it can be reused.
func putBuf(buf []byte) {
	size := cap(buf)
	if size < config.PooledSize {
		return
	}
	if c := buffers[size]; c != nil {
		c.Put(buf[:0])
	}
}

// getBuf gets a chunk from reuse pool or creates a new one if reuse failed.
func getBuf(size int) []byte {
	if size < config.PooledSize {
		return make([]byte, 0, size)
	}

	if c := buffers[size]; c != nil {
		v := c.Get()
		if v != nil {
			return v.([]byte)
		}
	}
	return make([]byte, 0, size)
}

// Buffer is a buffer optimized for serialization without extra copying.
type EJBuffer struct {

	// Buf is the current chunk that can be used for serialization.
	Buf []byte

	toPool []byte
	bufs   [][]byte
}

// EnsureSpace makes sure that the current chunk contains at least s free bytes,
// possibly creating a new chunk.
func (b *EJBuffer) EnsureSpace(s int) {
	if cap(b.Buf)-len(b.Buf) >= s {
		return
	}
	l := len(b.Buf)
	if l > 0 {
		if cap(b.toPool) != cap(b.Buf) {
			// Chunk was reallocated, toPool can be pooled.
			putBuf(b.toPool)
		}
		if cap(b.bufs) == 0 {
			b.bufs = make([][]byte, 0, 8)
		}
		b.bufs = append(b.bufs, b.Buf)
		l = cap(b.toPool) * 2
	} else {
		l = config.StartSize
	}

	if l > config.MaxSize {
		l = config.MaxSize
	}
	b.Buf = getBuf(l)
	b.toPool = b.Buf
}

// AppendBytes appends a byte slice to buffer.
func (b *EJBuffer) Write(data []byte) {
	for len(data) > 0 {
		if cap(b.Buf) == len(b.Buf) { // EnsureSpace won't be inlined.
			b.EnsureSpace(1)
		}

		sz := cap(b.Buf) - len(b.Buf)
		if sz > len(data) {
			sz = len(data)
		}

		b.Buf = append(b.Buf, data[:sz]...)
		data = data[sz:]
	}
}

// AppendByte appends a single byte to buffer.
func (b *EJBuffer) WriteByte(data byte) {
	if cap(b.Buf) == len(b.Buf) { // EnsureSpace won't be inlined.
		b.EnsureSpace(1)
	}
	b.Buf = append(b.Buf, data)
}

// AppendBytes appends a string to buffer.
func (b *EJBuffer) WriteString(data string) {
	for len(data) > 0 {
		if cap(b.Buf) == len(b.Buf) { // EnsureSpace won't be inlined.
			b.EnsureSpace(1)
		}

		sz := cap(b.Buf) - len(b.Buf)
		if sz > len(data) {
			sz = len(data)
		}

		b.Buf = append(b.Buf, data[:sz]...)
		data = data[sz:]
	}
}

// Size computes the size of a buffer by adding sizes of every chunk.
func (b *EJBuffer) Size() int {
	size := len(b.Buf)
	for _, buf := range b.bufs {
		size += len(buf)
	}
	return size
}

func (b *EJBuffer) Reset() {
	for _, buf := range b.bufs {
		putBuf(buf)
	}
	putBuf(b.toPool)

	b.bufs = nil
	b.Buf = nil
	b.toPool = nil
}
