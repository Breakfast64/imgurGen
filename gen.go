package main

import (
	"math/rand"
	"time"
)

type gen struct {
	rng    *rand.Rand
	buffer []byte
	cfg    userConfig
}

func createGen(config userConfig) gen {
	g := gen{
		rng: rand.New(
			rand.NewSource(
				time.Now().UnixNano(),
			),
		),
		cfg:    config,
		buffer: make([]byte, config.Length+1+len(config.Extension)),
	}
	g.buffer[0] = '/'
	return g
}

func (g *gen) next() string {
	path := g.buffer[1:]
	for i := 0; i < g.cfg.Length; i++ {
		path[i] = alphaNum[g.rng.Intn(len(alphaNum))]
	}
	ext := path[g.cfg.Length:]
	copy(ext, g.cfg.Extension)
	return string(g.buffer)
}
