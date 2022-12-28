package effects

import "github.com/faiface/beep"

// Gain amplifies the wrapped Streamer. The output of the wrapped Streamer gets multiplied by
// 1+Gain.
//
// Note that gain is not equivalent to the human perception of volume. Human perception of volume is
// roughly exponential, while gain only amplifies linearly.
type Gain[S beep.Size, P beep.Point[S]] struct {
	Streamer beep.Streamer[S, P]
	Gain     float64
}

// Stream streams the wrapped Streamer amplified by Gain.
func (g *Gain[S, P]) Stream(samples []P) (n int, ok bool) {
	n, ok = g.Streamer.Stream(samples)
	for i := range samples[:n] {
		samples[i][0] *= S(1 + g.Gain)
		samples[i][1] *= S(1 + g.Gain)
	}
	return n, ok
}

// Err propagates the wrapped Streamer's errors.
func (g *Gain[S, P]) Err() error {
	return g.Streamer.Err()
}
