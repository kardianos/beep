// Package speaker implements playback of beep.Streamer values through physical speakers.
package speaker

import (
	"sync"

	"github.com/faiface/beep"
	"github.com/hajimehoshi/oto"
	"github.com/pkg/errors"
)

type Player[S beep.Size, P beep.Point[S]] struct {
	mu      sync.Mutex
	mixer   beep.Mixer[S, P]
	samples []P
	buf     []byte
	context *oto.Context
	player  *oto.Player
	done    chan struct{}
}

// New initializes audio playback through speaker. Must be called before using this package.
//
// The bufferSize argument specifies the number of samples of the speaker's buffer. Bigger
// bufferSize means lower CPU usage and more reliable playback. Lower bufferSize means better
// responsiveness and less delay.
func New[S beep.Size, P beep.Point[S]](sampleRate beep.SampleRate, bufferSize int) (*Player[S, P], error) {
	p := &Player[S, P]{}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Close()

	p.mixer = beep.Mixer[S, P]{}

	numBytes := bufferSize * 4
	p.samples = make([]P, bufferSize)
	p.buf = make([]byte, numBytes)

	var err error
	p.context, err = oto.NewContext(int(sampleRate), 2, 2, numBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize speaker")
	}
	p.player = p.context.NewPlayer()

	p.done = make(chan struct{})

	go func() {
		for {
			select {
			default:
				p.update()
			case <-p.done:
				return
			}
		}
	}()

	return p, nil
}

// Close closes the playback and the driver. In most cases, there is certainly no need to call Close
// even when the program doesn't play anymore, because in properly set systems, the default mixer
// handles multiple concurrent processes. It's only when the default device is not a virtual but hardware
// device, that you'll probably want to manually manage the device from your application.
func (p *Player[S, P]) Close() {
	if p.player != nil {
		if p.done != nil {
			p.done <- struct{}{}
			p.done = nil
		}
		p.player.Close()
		p.context.Close()
		p.player = nil
	}
}

// Lock locks the speaker. While locked, speaker won't pull new data from the playing Streamers. Lock
// if you want to modify any currently playing Streamers to avoid race conditions.
//
// Always lock speaker for as little time as possible, to avoid playback glitches.
func (p *Player[S, P]) Lock() {
	p.mu.Lock()
}

// Unlock unlocks the speaker. Call after modifying any currently playing Streamer.
func (p *Player[S, P]) Unlock() {
	p.mu.Unlock()
}

// Play starts playing all provided Streamers through the speaker.
func (p *Player[S, P]) Play(s ...beep.Streamer[S, P]) {
	p.mu.Lock()
	p.mixer.Add(s...)
	p.mu.Unlock()
}

// Clear removes all currently playing Streamers from the speaker.
func (p *Player[S, P]) Clear() {
	p.mu.Lock()
	p.mixer.Clear()
	p.mu.Unlock()
}

// update pulls new data from the playing Streamers and sends it to the speaker. Blocks until the
// data is sent and started playing.
func (p *Player[S, P]) update() {
	p.mu.Lock()
	p.mixer.Stream(p.samples)
	p.mu.Unlock()

	for i := range p.samples {
		for c := range p.samples[i] {
			val := p.samples[i][c]
			if val < -1 {
				val = -1
			}
			if val > +1 {
				val = +1
			}
			valInt16 := int16(val * (1<<15 - 1))
			low := byte(valInt16)
			high := byte(valInt16 >> 8)
			p.buf[i*4+c*2+0] = low
			p.buf[i*4+c*2+1] = high
		}
	}

	p.player.Write(p.buf)
}
