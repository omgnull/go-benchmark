package main

import (
	"testing"
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
