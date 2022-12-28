package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
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

	ctrl := &beep.Ctrl[float64, [2]float64]{Streamer: beep.Loop[float64, [2]float64](-1, streamer), Paused: false}
	volume := &effects.Volume[float64, [2]float64]{
		Streamer: ctrl,
		Base:     2,
		Volume:   0,
		Silent:   false,
	}
	speedy := beep.ResampleRatio[float64, [2]float64](4, 1, volume)
	player.Play(speedy)

	for {
		fmt.Print("Press [ENTER] to pause/resume. ")
		fmt.Scanln()

		player.Lock()
		ctrl.Paused = !ctrl.Paused
		volume.Volume += 0.5
		speedy.SetRatio(speedy.Ratio() + 0.1)
		player.Unlock()
	}
}
