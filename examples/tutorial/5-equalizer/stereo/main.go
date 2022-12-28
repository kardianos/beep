package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
)

func noise() beep.Streamer[float64, [2]float64] {
	return beep.StreamerFunc[float64, [2]float64](func(samples [][2]float64) (n int, ok bool) {
		for i := range samples {
			samples[i][0] = rand.Float64()*2 - 1
			samples[i][1] = rand.Float64()*2 - 1
		}
		return len(samples), true
	})
}

func main() {
	sr := beep.SampleRate(44100)
	player, err := speaker.New[float64, [2]float64](sr, sr.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}
	defer player.Close()

	eq := effects.NewEqualizer[float64, [2]float64](noise(), sr, effects.StereoEqualizerSections[float64, [2]float64]{
		{
			Left:  effects.MonoEqualizerSection[float64, [2]float64]{F0: 200, Bf: 5, GB: 3, G0: 0, G: 8},
			Right: effects.MonoEqualizerSection[float64, [2]float64]{F0: 200, Bf: 5, GB: 3, G0: 0, G: -8},
		},
		{
			Left:  effects.MonoEqualizerSection[float64, [2]float64]{F0: 10000, Bf: 1000, GB: 3, G0: 0, G: 90},
			Right: effects.MonoEqualizerSection[float64, [2]float64]{F0: 10000, Bf: 1000, GB: 3, G0: 0, G: -90},
		},
	})

	player.Play(eq)
	select {}
}
