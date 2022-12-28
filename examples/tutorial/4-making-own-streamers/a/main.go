package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

func Noise[S beep.Size, P beep.Point[S]]() beep.Streamer[S, P] {
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

	player.Play(Noise[float64, [2]float64]())
	select {}
}
