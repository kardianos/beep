package effects

import (
	"math"

	"github.com/faiface/beep"
)

// Volume adjusts the volume of the wrapped Streamer in a human-natural way. Human's perception of
// volume is roughly logarithmic to gain and thus the natural way to adjust volume is exponential.
//
// Natural Base for the exponentiation is somewhere around 2. In order to adjust volume along
// decibells, pick 10 as the Base and set Volume to dB/10. However, adjusting volume along decibells
// is nowhere as natural as with bases around 2.
//
// Volume of 0 means no change. Negative Volume will decrease the perceived volume and positive will
// increase it.
//
// With exponential gain it's impossible to achieve the zero volume. When Silent field is set to
// true, the output is muted.
type Volume[S beep.Size, P beep.Point[S]] struct {
	Streamer beep.Streamer[S, P]
	Base     float64
	Volume   float64
	Silent   bool
}

// Stream streams the wrapped Streamer with volume adjusted according to Base, Volume and Silent
// fields.
func (v *Volume[S, P]) Stream(samples []P) (n int, ok bool) {
	n, ok = v.Streamer.Stream(samples)
	gain := 0.0
	if !v.Silent {
		gain = math.Pow(v.Base, v.Volume)
	}
	for i := range samples[:n] {
		samples[i][0] *= S(gain)
		samples[i][1] *= S(gain)
	}
	return n, ok
}

// Err propagates the wrapped Streamer's errors.
func (v *Volume[S, P]) Err() error {
	return v.Streamer.Err()
}
