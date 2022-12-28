// Package mp3 implements audio data decoding in MP3 format.
package mp3

import (
	"fmt"
	"io"

	"github.com/faiface/beep"
	gomp3 "github.com/hajimehoshi/go-mp3"
	"github.com/pkg/errors"
)

const (
	gomp3NumChannels   = 2
	gomp3Precision     = 2
	gomp3BytesPerFrame = gomp3NumChannels * gomp3Precision
)

// Decode takes a ReadCloser containing audio data in MP3 format and returns a StreamSeekCloser,
// which streams that audio. The Seek method will panic if rc is not io.Seeker.
//
// Do not close the supplied ReadSeekCloser, instead, use the Close method of the returned
// StreamSeekCloser when you want to release the resources.
func Decode[S beep.Size, P beep.Point[S]](rc io.ReadCloser) (s beep.StreamSeekCloser[S, P], format beep.Format[S, P], err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "mp3")
		}
	}()
	d, err := gomp3.NewDecoder(rc)
	if err != nil {
		return nil, beep.Format[S, P]{}, err
	}
	format = beep.Format[S, P]{
		SampleRate:  beep.SampleRate(d.SampleRate()),
		NumChannels: gomp3NumChannels,
		Precision:   gomp3Precision,
	}
	return &decoder[S, P]{rc, d, format, 0, nil}, format, nil
}

type decoder[S beep.Size, P beep.Point[S]] struct {
	closer io.Closer
	d      *gomp3.Decoder
	f      beep.Format[S, P]
	pos    int
	err    error
}

func (d *decoder[S, P]) Stream(samples []P) (n int, ok bool) {
	if d.err != nil {
		return 0, false
	}
	var tmp [gomp3BytesPerFrame]byte
	for i := range samples {
		dn, err := d.d.Read(tmp[:])
		if dn == len(tmp) {
			samples[i], _ = d.f.DecodeSigned(tmp[:])
			d.pos += dn
			n++
			ok = true
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			d.err = errors.Wrap(err, "mp3")
			break
		}
	}
	return n, ok
}

func (d *decoder[S, P]) Err() error {
	return d.err
}

func (d *decoder[S, P]) Len() int {
	return int(d.d.Length()) / gomp3BytesPerFrame
}

func (d *decoder[S, P]) Position() int {
	return d.pos / gomp3BytesPerFrame
}

func (d *decoder[S, P]) Seek(p int) error {
	if p < 0 || d.Len() < p {
		return fmt.Errorf("mp3: seek position %v out of range [%v, %v]", p, 0, d.Len())
	}
	_, err := d.d.Seek(int64(p)*gomp3BytesPerFrame, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, "mp3")
	}
	d.pos = p * gomp3BytesPerFrame
	return nil
}

func (d *decoder[S, P]) Close() error {
	err := d.closer.Close()
	if err != nil {
		return errors.Wrap(err, "mp3")
	}
	return nil
}
