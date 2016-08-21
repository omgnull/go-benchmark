package main

import (
	"testing"

	"github.com/omgnull/go-benchmark/buffer"
)

func BenchmarkGenericBuf(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GenericBuf()
		}
	})
}

func BenchmarkGenericStackBuf(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GenericStackBuf()
		}
	})
}

func BenchmarkAllocBuf(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			AllocBuf()
		}
	})
}

func BenchmarkSyncPoolBuf(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			SyncBuf()
		}
	})
}

func BenchmarkBpoolPoolBuf(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			BpoolBuf()
		}
	})
}

func BenchmarkByteBufferPoolBuf(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			BBpoolBuf()
		}
	})
}

func BenchmarkEasyJsonBuffer(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			EasyJsonBuf()
		}
	})
}

func BenchmarkEasyJsonBuffer_OptimizedConfig(b *testing.B) {
	buffer.Init(buffer.PoolConfig{
		StartSize:  2048,
		PooledSize: 2048,
		MaxSize:    32768,
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			EasyJsonBuf()
		}
	})
}
