package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
)

func noise[S beep.Size, P beep.Point[S]]() beep.Streamer[S, P] {
	return beep.StreamerFunc[S, P](func(samples []P) (n int, ok bool) {
		for i := range samples {
			samples[i][0] = S(rand.Float64()*2 - 1)
			samples[i][1] = S(rand.Float64()*2 - 1)
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

	eq := effects.NewEqualizer[float64, [2]float64](noise[float64, [2]float64](), sr, effects.MonoEqualizerSections[float64, [2]float64]{
		{F0: 200, Bf: 5, GB: 3, G0: 0, G: 8},
		{F0: 250, Bf: 5, GB: 3, G0: 0, G: 10},
		{F0: 300, Bf: 5, GB: 3, G0: 0, G: 12},
		{F0: 350, Bf: 5, GB: 3, G0: 0, G: 14},
		{F0: 10000, Bf: 8000, GB: 3, G0: 0, G: -100},
	})

	player.Play(eq)
	select {}
}
