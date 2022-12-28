package main

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func main() {
	f, err := os.Open("../Lame_Drivers_-_01_-_Frozen_Egg.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode[float64, [2]float64](f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	sr := format.SampleRate * 2
	player, err := speaker.New[float64, [2]float64](sr, sr.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}
	defer player.Close()

	resampled := beep.Resample[float64, [2]float64](4, format.SampleRate, sr, streamer)

	done := make(chan bool)
	player.Play(beep.Seq[float64, [2]float64](resampled, beep.Callback[float64, [2]float64](func() {
		done <- true
	})))

	<-done
}
