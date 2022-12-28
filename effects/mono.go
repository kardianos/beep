package effects

import "github.com/faiface/beep"

// Mono converts the wrapped Streamer to a mono buffer
// by downmixing the left and right channels together.
//
// The returned Streamer propagates s's errors through Err.
func Mono[S beep.Size, P beep.Point[S]](s beep.Streamer[S, P]) beep.Streamer[S, P] {
	return &mono[S, P]{s}
}

type mono[S beep.Size, P beep.Point[S]] struct {
	Streamer beep.Streamer[S, P]
}

func (m *mono[S, P]) Stream(samples []P) (n int, ok bool) {
	n, ok = m.Streamer.Stream(samples)
	for i := range samples[:n] {
		mix := (samples[i][0] + samples[i][1]) / 2
		samples[i][0], samples[i][1] = mix, mix
	}
	return n, ok
}

func (m *mono[S, P]) Err() error {
	return m.Streamer.Err()
}
