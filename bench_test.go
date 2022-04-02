package main

import "testing"

func BenchmarkGen(b *testing.B) {
	gen := createGen(defCfg)
	for i := 0; i < b.N; i++ {
		gen.next()
	}
}
