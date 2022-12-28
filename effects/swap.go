package effects

import "github.com/faiface/beep"

// Swap swaps the left and right channel of the wrapped Streamer.
//
// The returned Streamer propagates s's errors through Err.
func Swap[S beep.Size, P beep.Point[S]](s beep.Streamer[S, P]) beep.Streamer[S, P] {
	return &swap[S, P]{s}
}

type swap[S beep.Size, P beep.Point[S]] struct {
	Streamer beep.Streamer[S, P]
}

func (s *swap[S, P]) Stream(samples []P) (n int, ok bool) {
	n, ok = s.Streamer.Stream(samples)
	for i := range samples[:n] {
		samples[i][0], samples[i][1] = samples[i][1], samples[i][0]
	}
	return n, ok
}

func (s *swap[S, P]) Err() error {
	return s.Streamer.Err()
}
