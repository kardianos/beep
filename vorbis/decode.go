// Package vorbis implements audio data decoding in oggvorbis format.
package vorbis

import (
	"io"

	"github.com/faiface/beep"
	"github.com/jfreymuth/oggvorbis"
	"github.com/pkg/errors"
)

const (
	govorbisNumChannels = 2
	govorbisPrecision   = 2
)

// Decode takes a ReadCloser containing audio data in ogg/vorbis format and returns a StreamSeekCloser,
// which streams that audio. The Seek method will panic if rc is not io.Seeker.
//
// Do not close the supplied ReadSeekCloser, instead, use the Close method of the returned
// StreamSeekCloser when you want to release the resources.
func Decode[S beep.Size, P beep.Point[S]](rc io.ReadCloser) (s beep.StreamSeekCloser[S, P], format beep.Format[S, P], err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "ogg/vorbis")
		}
	}()
	d, err := oggvorbis.NewReader(rc)
	if err != nil {
		return nil, beep.Format[S, P]{}, err
	}
	format = beep.Format[S, P]{
		SampleRate:  beep.SampleRate(d.SampleRate()),
		NumChannels: govorbisNumChannels,
		Precision:   govorbisPrecision,
	}
	return &decoder[S, P]{rc, d, format, nil}, format, nil
}

type decoder[S beep.Size, P beep.Point[S]] struct {
	closer io.Closer
	d      *oggvorbis.Reader
	f      beep.Format[S, P]
	err    error
}

func (d *decoder[S, P]) Stream(samples []P) (n int, ok bool) {
	if d.err != nil {
		return 0, false
	}
	var tmp [2]float32
	for i := range samples {
		dn, err := d.d.Read(tmp[:])
		if dn == 2 {
			samples[i][0], samples[i][1] = S(tmp[0]), S(tmp[1])
			n++
			ok = true
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			d.err = errors.Wrap(err, "ogg/vorbis")
			break
		}
	}
	return n, ok
}

func (d *decoder[S, P]) Err() error {
	return d.err
}

func (d *decoder[S, P]) Len() int {
	return int(d.d.Length())
}

func (d *decoder[S, P]) Position() int {
	return int(d.d.Position())
}

func (d *decoder[S, P]) Seek(p int) error {
	err := d.d.SetPosition(int64(p))
	if err != nil {
		return errors.Wrap(err, "ogg/vorbis")
	}
	return nil
}

func (d *decoder[S, P]) Close() error {
	err := d.closer.Close()
	if err != nil {
		return errors.Wrap(err, "ogg/vorbis")
	}
	return nil
}
