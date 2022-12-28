package flac

import (
	"fmt"
	"io"

	"github.com/faiface/beep"
	"github.com/mewkiz/flac"
	"github.com/pkg/errors"
)

// Decode takes a Reader containing audio data in FLAC format and returns a StreamSeekCloser,
// which streams that audio. The Seek method will panic if r is not io.Seeker.
//
// Do not close the supplied Reader, instead, use the Close method of the returned
// StreamSeekCloser when you want to release the resources.
func Decode[S beep.Size, P beep.Point[S]](r io.Reader) (s beep.StreamSeekCloser[S, P], format beep.Format[S, P], err error) {
	d := decoder[S, P]{r: r}
	defer func() { // hacky way to always close r if an error occurred
		if closer, ok := d.r.(io.Closer); ok {
			if err != nil {
				closer.Close()
			}
		}
	}()

	rs, seeker := r.(io.ReadSeeker)
	if seeker {
		d.stream, err = flac.NewSeek(rs)
		d.seekEnabled = true
	} else {
		d.stream, err = flac.New(r)
	}

	if err != nil {
		return nil, beep.Format[S, P]{}, errors.Wrap(err, "flac")
	}
	format = beep.Format[S, P]{
		SampleRate:  beep.SampleRate(d.stream.Info.SampleRate),
		NumChannels: int(d.stream.Info.NChannels),
		Precision:   int(d.stream.Info.BitsPerSample / 8),
	}
	return &d, format, nil
}

type decoder[S beep.Size, P beep.Point[S]] struct {
	r           io.Reader
	stream      *flac.Stream
	buf         []P
	pos         int
	err         error
	seekEnabled bool
}

func (d *decoder[S, P]) Stream(samples []P) (n int, ok bool) {
	if d.err != nil {
		return 0, false
	}
	// Copy samples from buffer.
	j := 0
	for i := range samples {
		if j >= len(d.buf) {
			// refill buffer.
			if err := d.refill(); err != nil {
				d.err = err
				d.pos += n
				return n, n > 0
			}
			j = 0
		}
		samples[i] = d.buf[j]
		j++
		n++
	}
	d.buf = d.buf[j:]
	d.pos += n
	return n, true
}

// refill decodes audio samples to fill the decode buffer.
func (d *decoder[S, P]) refill() error {
	// Empty buffer.
	d.buf = d.buf[:0]
	// Parse audio frame.
	frame, err := d.stream.ParseNext()
	if err != nil {
		return err
	}
	// Expand buffer size if needed.
	n := len(frame.Subframes[0].Samples)
	if cap(d.buf) < n {
		d.buf = make([]P, n)
	} else {
		d.buf = d.buf[:n]
	}
	// Decode audio samples.
	bps := d.stream.Info.BitsPerSample
	nchannels := d.stream.Info.NChannels
	s := 1 << (bps - 1)
	q := 1 / S(s)
	switch {
	case bps == 8 && nchannels == 1:
		for i := 0; i < n; i++ {
			d.buf[i][0] = S(int8(frame.Subframes[0].Samples[i])) * q
			d.buf[i][1] = S(int8(frame.Subframes[0].Samples[i])) * q
		}
	case bps == 16 && nchannels == 1:
		for i := 0; i < n; i++ {
			d.buf[i][0] = S(int16(frame.Subframes[0].Samples[i])) * q
			d.buf[i][1] = S(int16(frame.Subframes[0].Samples[i])) * q
		}
	case bps == 24 && nchannels == 1:
		for i := 0; i < n; i++ {
			d.buf[i][0] = S(int32(frame.Subframes[0].Samples[i])) * q
			d.buf[i][1] = S(int32(frame.Subframes[0].Samples[i])) * q
		}
	case bps == 8 && nchannels >= 2:
		for i := 0; i < n; i++ {
			d.buf[i][0] = S(int8(frame.Subframes[0].Samples[i])) * q
			d.buf[i][1] = S(int8(frame.Subframes[1].Samples[i])) * q
		}
	case bps == 16 && nchannels >= 2:
		for i := 0; i < n; i++ {
			d.buf[i][0] = S(int16(frame.Subframes[0].Samples[i])) * q
			d.buf[i][1] = S(int16(frame.Subframes[1].Samples[i])) * q
		}
	case bps == 24 && nchannels >= 2:
		for i := 0; i < n; i++ {
			d.buf[i][0] = S(frame.Subframes[0].Samples[i]) * q
			d.buf[i][1] = S(frame.Subframes[1].Samples[i]) * q
		}
	default:
		panic(fmt.Errorf("support for %d bits-per-sample and %d channels combination not yet implemented", bps, nchannels))
	}
	return nil
}

func (d *decoder[S, P]) Err() error {
	return d.err
}

func (d *decoder[S, P]) Len() int {
	return int(d.stream.Info.NSamples)
}

func (d *decoder[S, P]) Position() int {
	return d.pos
}

// p represents flac sample num perhaps?
func (d *decoder[S, P]) Seek(p int) error {
	if !d.seekEnabled {
		return errors.New("flac.decoder.Seek: not enabled")
	}

	pos, err := d.stream.Seek(uint64(p))
	d.pos = int(pos)
	return err
}

func (d *decoder[S, P]) Close() error {
	if closer, ok := d.r.(io.Closer); ok {
		err := closer.Close()
		if err != nil {
			return errors.Wrap(err, "flac")
		}
	}
	return nil
}
