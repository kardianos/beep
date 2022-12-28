package main

import (
	"fmt"
	"os"
	"time"
	"unicode"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gdamore/tcell"
)

func drawTextLine(screen tcell.Screen, x, y int, s string, style tcell.Style) {
	for _, r := range s {
		screen.SetContent(x, y, r, nil, style)
		x++
	}
}

type audioPanel[S beep.Size, P beep.Point[S]] struct {
	player     *speaker.Player[S, P]
	sampleRate beep.SampleRate
	streamer   beep.StreamSeeker[S, P]
	ctrl       *beep.Ctrl[S, P]
	resampler  *beep.Resampler[S, P]
	volume     *effects.Volume[S, P]
}

func newAudioPanel[S beep.Size, P beep.Point[S]](player *speaker.Player[S, P], sampleRate beep.SampleRate, streamer beep.StreamSeeker[S, P]) *audioPanel[S, P] {
	ctrl := &beep.Ctrl[S, P]{Streamer: beep.Loop(-1, streamer)}
	resampler := beep.ResampleRatio[S, P](4, 1, ctrl)
	volume := &effects.Volume[S, P]{Streamer: resampler, Base: 2}
	return &audioPanel[S, P]{player, sampleRate, streamer, ctrl, resampler, volume}
}

func (ap *audioPanel[S, P]) play() {
	ap.player.Play(ap.volume)
}

func (ap *audioPanel[S, P]) draw(screen tcell.Screen) {
	mainStyle := tcell.StyleDefault.
		Background(tcell.NewHexColor(0x473437)).
		Foreground(tcell.NewHexColor(0xD7D8A2))
	statusStyle := mainStyle.
		Foreground(tcell.NewHexColor(0xDDC074)).
		Bold(true)

	screen.Fill(' ', mainStyle)

	drawTextLine(screen, 0, 0, "Welcome to the Speedy Player!", mainStyle)
	drawTextLine(screen, 0, 1, "Press [ESC] to quit.", mainStyle)
	drawTextLine(screen, 0, 2, "Press [SPACE] to pause/resume.", mainStyle)
	drawTextLine(screen, 0, 3, "Use keys in (?/?) to turn the buttons.", mainStyle)

	ap.player.Lock()
	position := ap.sampleRate.D(ap.streamer.Position())
	length := ap.sampleRate.D(ap.streamer.Len())
	volume := ap.volume.Volume
	speed := ap.resampler.Ratio()
	ap.player.Unlock()

	positionStatus := fmt.Sprintf("%v / %v", position.Round(time.Second), length.Round(time.Second))
	volumeStatus := fmt.Sprintf("%.1f", volume)
	speedStatus := fmt.Sprintf("%.3fx", speed)

	drawTextLine(screen, 0, 5, "Position (Q/W):", mainStyle)
	drawTextLine(screen, 16, 5, positionStatus, statusStyle)

	drawTextLine(screen, 0, 6, "Volume   (A/S):", mainStyle)
	drawTextLine(screen, 16, 6, volumeStatus, statusStyle)

	drawTextLine(screen, 0, 7, "Speed    (Z/X):", mainStyle)
	drawTextLine(screen, 16, 7, speedStatus, statusStyle)
}

func (ap *audioPanel[S, P]) handle(event tcell.Event) (changed, quit bool) {
	switch event := event.(type) {
	case *tcell.EventKey:
		if event.Key() == tcell.KeyESC {
			return false, true
		}

		if event.Key() != tcell.KeyRune {
			return false, false
		}

		switch unicode.ToLower(event.Rune()) {
		case ' ':
			ap.player.Lock()
			ap.ctrl.Paused = !ap.ctrl.Paused
			ap.player.Unlock()
			return false, false

		case 'q', 'w':
			ap.player.Lock()
			newPos := ap.streamer.Position()
			if event.Rune() == 'q' {
				newPos -= ap.sampleRate.N(time.Second)
			}
			if event.Rune() == 'w' {
				newPos += ap.sampleRate.N(time.Second)
			}
			if newPos < 0 {
				newPos = 0
			}
			if newPos >= ap.streamer.Len() {
				newPos = ap.streamer.Len() - 1
			}
			if err := ap.streamer.Seek(newPos); err != nil {
				report(err)
			}
			ap.player.Unlock()
			return true, false

		case 'a':
			ap.player.Lock()
			ap.volume.Volume -= 0.1
			ap.player.Unlock()
			return true, false

		case 's':
			ap.player.Lock()
			ap.volume.Volume += 0.1
			ap.player.Unlock()
			return true, false

		case 'z':
			ap.player.Lock()
			ap.resampler.SetRatio(ap.resampler.Ratio() * 15 / 16)
			ap.player.Unlock()
			return true, false

		case 'x':
			ap.player.Lock()
			ap.resampler.SetRatio(ap.resampler.Ratio() * 16 / 15)
			ap.player.Unlock()
			return true, false
		}
	}
	return false, false
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s song.mp3\n", os.Args[0])
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		report(err)
	}
	streamer, format, err := mp3.Decode[float64, [2]float64](f)
	if err != nil {
		report(err)
	}
	defer streamer.Close()

	player, err := speaker.New[float64, [2]float64](format.SampleRate, format.SampleRate.N(time.Second/30))
	if err != nil {
		report(err)
	}
	defer player.Close()

	screen, err := tcell.NewScreen()
	if err != nil {
		report(err)
	}
	err = screen.Init()
	if err != nil {
		report(err)
	}
	defer screen.Fini()

	ap := newAudioPanel[float64, [2]float64](player, format.SampleRate, streamer)

	screen.Clear()
	ap.draw(screen)
	screen.Show()

	ap.play()

	seconds := time.Tick(time.Second)
	events := make(chan tcell.Event)
	go func() {
		for {
			events <- screen.PollEvent()
		}
	}()

loop:
	for {
		select {
		case event := <-events:
			changed, quit := ap.handle(event)
			if quit {
				break loop
			}
			if changed {
				screen.Clear()
				ap.draw(screen)
				screen.Show()
			}
		case <-seconds:
			screen.Clear()
			ap.draw(screen)
			screen.Show()
		}
	}
}

func report(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
