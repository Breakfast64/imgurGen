package main

import "time"

type Progress struct {
	States   []string
	Interval time.Duration
	cursor   int
	before   time.Time
}

func (p *Progress) Next() (out string, same bool) {
	now := time.Now()
	elapsed := now.Sub(p.before)
	if elapsed >= p.Interval {
		p.cursor++
		p.before = now
		same = false
	} else {
		same = true
	}

	idx := p.cursor % len(p.States)
	out = p.States[idx]
	return
}
