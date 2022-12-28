package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func main() {
	f, err := os.Open("../Miami_Slice_-_04_-_Step_Into_Me.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode[float64, [2]float64](f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	player, err := speaker.New[float64, [2]float64](format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}
	defer player.Close()

	done := make(chan bool)
	player.Play(beep.Seq[float64, [2]float64](streamer, beep.Callback[float64, [2]float64](func() {
		done <- true
	})))

	for {
		select {
		case <-done:
			return
		case <-time.After(time.Second):
			player.Lock()
			fmt.Println(format.SampleRate.D(streamer.Position()).Round(time.Second))
			player.Unlock()
		}
	}
}
