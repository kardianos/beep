package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/generators"
	"github.com/faiface/beep/speaker"
)

func usage() {
	fmt.Printf("usage: %s freq\n", os.Args[0])
	fmt.Println("where freq must be a float between 1 and 24000")
	fmt.Println("24000 because samplerate of 48000 is hardcoded")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	f, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		usage()
		return
	}

	sr := beep.SampleRate(48000)
	player, err := speaker.New[float64, [2]float64](sr, 4800)
	if err != nil {
		panic(err)
	}
	defer player.Close()

	sine, err := generators.SineTone[float64, [2]float64](sr, f)
	if err != nil {
		panic(err)
	}

	triangle, err := generators.TriangleTone[float64, [2]float64](sr, f)
	if err != nil {
		panic(err)
	}

	square, err := generators.SquareTone[float64, [2]float64](sr, f)
	if err != nil {
		panic(err)
	}

	sawtooth, err := generators.SawtoothTone[float64, [2]float64](sr, f)
	if err != nil {
		panic(err)
	}

	sawtoothReversed, err := generators.SawtoothToneReversed[float64, [2]float64](sr, f)
	if err != nil {
		panic(err)
	}

	// Play 2 seconds of each tone
	two := sr.N(2 * time.Second)

	ch := make(chan struct{})
	sounds := []beep.Streamer[float64, [2]float64]{
		beep.Callback[float64, [2]float64](print("sine")),
		beep.Take(two, sine),
		beep.Callback[float64, [2]float64](print("triangle")),
		beep.Take(two, triangle),
		beep.Callback[float64, [2]float64](print("square")),
		beep.Take(two, square),
		beep.Callback[float64, [2]float64](print("sawtooth")),
		beep.Take(two, sawtooth),
		beep.Callback[float64, [2]float64](print("sawtooth reversed")),
		beep.Take(two, sawtoothReversed),
		beep.Callback[float64, [2]float64](func() {
			ch <- struct{}{}
		}),
	}
	player.Play(beep.Seq(sounds...))

	<-ch
}

// a simple clousure to wrap fmt.Println
func print(s string) func() {
	return func() {
		fmt.Println(s)
	}
}
