package main

import (
	"github.com/valyala/fastrand"
)

type gen struct {
	rng    fastrand.RNG
	buffer []byte
	cfg    userConfig
}

func createGen(config userConfig) gen {

	g := gen{
		cfg:    config,
		buffer: make([]byte, config.Length+1+len(config.Extension)),
	}
	g.buffer[0] = '/'
	return g
}

func (g *gen) next() []byte {
	path := g.buffer[1:]
	for i := 0; i < g.cfg.Length; i++ {
		pathIdx := int(g.rng.Uint32()) % len(alphaNum)
		path[i] = alphaNum[pathIdx]
	}
	ext := path[g.cfg.Length:]
	copy(ext, g.cfg.Extension)
	return g.buffer
}
