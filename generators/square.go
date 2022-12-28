package generators

import (
	"errors"
	"math"

	"github.com/faiface/beep"
)

type squareGenerator[S beep.Size, P beep.Point[S]] struct {
	dt float64
	t  float64
}

// Creates a streamer which will procude an infinite square wave with the given frequency.
// use other wrappers of this package to change amplitude or add time limit.
// sampleRate must be at least two times grater then frequency, otherwise this function will return an error.
func SquareTone[S beep.Size, P beep.Point[S]](sr beep.SampleRate, freq float64) (beep.Streamer[S, P], error) {
	dt := freq / float64(sr)

	if dt >= 1.0/2.0 {
		return nil, errors.New("faiface square tone generator: samplerate must be at least 2 times grater then frequency")
	}

	return &squareGenerator[S, P]{dt, 0}, nil
}

func (g *squareGenerator[S, P]) Stream(samples []P) (n int, ok bool) {
	for i := range samples {
		if g.t < 0.5 {
			samples[i][0] = 1.0
			samples[i][1] = 1.0
		} else {
			samples[i][0] = -1.0
			samples[i][1] = -1.0
		}
		_, g.t = math.Modf(g.t + g.dt)
	}

	return len(samples), true
}

func (*squareGenerator[S, P]) Err() error {
	return nil
}
