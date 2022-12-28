package generators

import (
	"errors"
	"math"

	"github.com/faiface/beep"
)

type sawGenerator[S beep.Size, P beep.Point[S]] struct {
	dt float64
	t  float64

	reverse bool
}

// Creates a streamer which will procude an infinite sawtooth wave with the given frequency.
// use other wrappers of this package to change amplitude or add time limit.
// sampleRate must be at least two times grater then frequency, otherwise this function will return an error.
func SawtoothTone[S beep.Size, P beep.Point[S]](sr beep.SampleRate, freq float64) (beep.Streamer[S, P], error) {
	dt := freq / float64(sr)

	if dt >= 1.0/2.0 {
		return nil, errors.New("faiface sawtooth tone generator: samplerate must be at least 2 times grater then frequency")
	}

	return &sawGenerator[S, P]{dt, 0, false}, nil
}

// Creates a streamer which will procude an infinite sawtooth tone with the given frequency.
// sawtooth is reversed so the slope is negative.
// use other wrappers of this package to change amplitude or add time limit.
// sampleRate must be at least two times grater then frequency, otherwise this function will return an error.
func SawtoothToneReversed[S beep.Size, P beep.Point[S]](sr beep.SampleRate, freq float64) (beep.Streamer[S, P], error) {
	dt := freq / float64(sr)

	if dt >= 1.0/2.0 {
		return nil, errors.New("faiface triangle tone generator: samplerate must be at least 2 times grater then frequency")
	}

	return &sawGenerator[S, P]{dt, 0, true}, nil
}

func (g *sawGenerator[S, P]) Stream(samples []P) (n int, ok bool) {
	if g.reverse {
		for i := range samples {
			samples[i][0] = S(2.0*(1-g.t) - 1)
			samples[i][1] = S(2.0*(1-g.t) - 1)
			_, g.t = math.Modf(g.t + g.dt)
		}
	} else {
		for i := range samples {
			samples[i][0] = S(2.0*g.t - 1.0)
			samples[i][1] = S(2.0*g.t - 1.0)
			_, g.t = math.Modf(g.t + g.dt)
		}
	}

	return len(samples), true
}

func (*sawGenerator[S, P]) Err() error {
	return nil
}
