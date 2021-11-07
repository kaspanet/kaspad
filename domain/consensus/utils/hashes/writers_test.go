package hashes

import (
	"math/rand"
	"testing"
)

func BenchmarkNewBlockHashWriterSmall(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	var someBytes [32]byte
	r.Read(someBytes[:])
	for i := 0; i < b.N; i++ {
		hasher := NewBlockHashWriter()
		hasher.InfallibleWrite(someBytes[:])
		hasher.Finalize()
	}
}

func BenchmarkNewBlockHashWriterBig(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	var someBytes [1024]byte
	r.Read(someBytes[:])
	for i := 0; i < b.N; i++ {
		hasher := NewBlockHashWriter()
		hasher.InfallibleWrite(someBytes[:])
		hasher.Finalize()
	}
}

func BenchmarkNewHeavyHashWriterSmall(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	var someBytes [32]byte
	r.Read(someBytes[:])
	for i := 0; i < b.N; i++ {
		hasher := NewHeavyHashWriter()
		hasher.InfallibleWrite(someBytes[:])
		hasher.Finalize()
	}
}

func BenchmarkNewHeavyHashWriterBig(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	var someBytes [1024]byte
	r.Read(someBytes[:])
	for i := 0; i < b.N; i++ {
		hasher := NewHeavyHashWriter()
		hasher.InfallibleWrite(someBytes[:])
		hasher.Finalize()
	}
}
