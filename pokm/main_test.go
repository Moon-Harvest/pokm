package main

import (
	"testing"
)

func BenchmarkLookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for id := 1; id < 1025; id++ {
			getPokemonData(uint16(id))
		}
	}
}
