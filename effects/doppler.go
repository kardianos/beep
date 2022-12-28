package effects

import "github.com/faiface/beep"

// Doppler simulates a "sound at a distance". If the sound starts at a far distance,
// it'll take some time to reach the ears of the listener.
//
// The distance of the sound can change dynamically. Doppler adjusts the density of
// the sound (affecting its speed) to remain consistent with the distance. This is called
// the Doppler effect.
//
// The arguments are:
//
//	quality:         the quality of the underlying resampler (1 or 2 is usually okay)
//	samplesPerMeter: sample rate / speed of sound
//	s:               the source streamer
//	distance:        a function to calculate the current distance; takes number of
//	                 samples Doppler wants to stream at the moment
//
// This function is experimental and may change any time!
func Doppler[S beep.Size, P beep.Point[S]](quality int, samplesPerMeter float64, s beep.Streamer[S, P], distance func(delta int) float64) beep.Streamer[S, P] {
	return &doppler[S, P]{
		r:               beep.ResampleRatio(quality, 1, s),
		distance:        distance,
		space:           make([]P, int(distance(0)*samplesPerMeter)),
		samplesPerMeter: samplesPerMeter,
	}
}

type doppler[S beep.Size, P beep.Point[S]] struct {
	r               *beep.Resampler[S, P]
	distance        func(delta int) float64
	space           []P
	samplesPerMeter float64
}

func (d *doppler[S, P]) Stream(samples []P) (n int, ok bool) {
	distance := d.distance(len(samples))
	currentSpaceLen := int(distance * d.samplesPerMeter)
	difference := currentSpaceLen - len(d.space)

	d.r.SetRatio(float64(len(samples)) / float64(len(samples)+difference))

	d.space = append(d.space, make([]P, len(samples)+difference)...)
	rn, _ := d.r.Stream(d.space[len(d.space)-len(samples)-difference:])
	d.space = d.space[:len(d.space)-len(samples)-difference+rn]
	for i := len(d.space) - rn; i < len(d.space); i++ {
		d.space[i][0] /= S(distance * distance)
		d.space[i][1] /= S(distance * distance)
	}

	if len(d.space) == 0 {
		return 0, false
	}
	n = copy(samples, d.space)
	d.space = d.space[n:]
	return n, true
}

func (d *doppler[S, P]) Err() error {
	return d.r.Err()
}
