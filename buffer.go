package beep

import (
	"fmt"
	"math"
	"time"
)

// SampleRate is the number of samples per second.
type SampleRate int

// D returns the duration of n samples.
func (sr SampleRate) D(n int) time.Duration {
	return time.Second * time.Duration(n) / time.Duration(sr)
}

// N returns the number of samples that last for d duration.
func (sr SampleRate) N(d time.Duration) int {
	return int(d * time.Duration(sr) / time.Second)
}

// Format is the format of a Buffer or another audio source.
type Format[S Size, P Point[S]] struct {
	// SampleRate is the number of samples per second.
	SampleRate SampleRate

	// NumChannels is the number of channels. The value of 1 is mono, the value of 2 is stereo.
	// The samples should always be interleaved.
	NumChannels int

	// Precision is the number of bytes used to encode a single sample. Only values up to 6 work
	// well, higher values loose precision due to floating point numbers.
	Precision int
}

// Width returns the number of bytes per one frame (samples in all channels).
//
// This is equal to f.NumChannels * f.Precision.
func (f Format[S, P]) Width() int {
	return f.NumChannels * f.Precision
}

// EncodeSigned encodes a single sample in f.Width() bytes to p in signed format.
func (f Format[S, P]) EncodeSigned(p []byte, sample P) (n int) {
	return f.encode(true, p, sample)
}

// EncodeUnsigned encodes a single sample in f.Width() bytes to p in unsigned format.
func (f Format[S, P]) EncodeUnsigned(p []byte, sample P) (n int) {
	return f.encode(false, p, sample)
}

// DecodeSigned decodes a single sample encoded in f.Width() bytes from p in signed format.
func (f Format[S, P]) DecodeSigned(p []byte) (sample P, n int) {
	return f.decode(true, p)
}

// DecodeUnsigned decodes a single sample encoded in f.Width() bytes from p in unsigned format.
func (f Format[S, P]) DecodeUnsigned(p []byte) (sample P, n int) {
	return f.decode(false, p)
}

func (f Format[S, P]) encode(signed bool, p []byte, sample P) (n int) {
	switch {
	case f.NumChannels == 1:
		var x S
		ct := sample.Count()
		if ct == 1 {
			x = sample.Get(0)
		} else {
			for _, v := range sample.Slice() {
				x += v
			}
			x = norm(x / S(ct))
		}
		p = p[encodeFloat(signed, f.Precision, p, x):]
	case f.NumChannels >= 2:
		for _, v := range sample.Slice() {
			x := norm(v)
			p = p[encodeFloat(signed, f.Precision, p, x):]
		}
		for c := sample.Count(); c < f.NumChannels; c++ {
			p = p[encodeFloat[S](signed, f.Precision, p, 0):]
		}
	default:
		panic(fmt.Errorf("format: encode: invalid number of channels: %d", f.NumChannels))
	}
	return f.Width()
}

func (f Format[S, P]) decode(signed bool, p []byte) (sample P, n int) {
	switch {
	case f.NumChannels == 1:
		x, _ := decodeFloat[S](signed, f.Precision, p)
		var xx P
		for i := range xx.Slice() {
			xx = xx.Set(i, x).(P)
		}
		return xx, f.Width()
	case f.NumChannels >= 2:
		sl := sample.Slice()
		for c := range sl {
			x, n := decodeFloat[S](signed, f.Precision, p)
			sample = sample.Set(c, x).(P)
			p = p[n:]
		}
		for c := sample.Count(); c < f.NumChannels; c++ {
			_, n := decodeFloat[S](signed, f.Precision, p)
			p = p[n:]
		}
		return sample, f.Width()
	default:
		panic(fmt.Errorf("format: decode: invalid number of channels: %d", f.NumChannels))
	}
}

func encodeFloat[S Size](signed bool, precision int, p []byte, x S) (n int) {
	var xUint64 uint64
	if signed {
		xUint64 = floatToSigned(precision, x)
	} else {
		xUint64 = floatToUnsigned(precision, x)
	}
	for i := 0; i < precision; i++ {
		p[i] = byte(xUint64)
		xUint64 >>= 8
	}
	return precision
}

func decodeFloat[S Size](signed bool, precision int, p []byte) (x S, n int) {
	var xUint64 uint64
	for i := precision - 1; i >= 0; i-- {
		xUint64 <<= 8
		xUint64 += uint64(p[i])
	}
	if signed {
		return signedToFloat[S](precision, xUint64), precision
	}
	return unsignedToFloat[S](precision, xUint64), precision
}

func floatToSigned[S Size](precision int, x S) uint64 {
	if x < 0 {
		compl := uint64(-float64(x) * (math.Exp2(float64(precision)*8-1) - 1))
		return uint64(1<<uint(precision*8)) - compl
	}
	return uint64(float64(x) * (math.Exp2(float64(precision)*8-1) - 1))
}

func floatToUnsigned[S Size](precision int, x S) uint64 {
	return uint64((float64(x) + 1) / 2 * (math.Exp2(float64(precision)*8) - 1))
}

func signedToFloat[S Size](precision int, xUint64 uint64) S {
	if xUint64 >= 1<<uint(precision*8-1) {
		compl := 1<<uint(precision*8) - xUint64
		return -S(int64(compl)) / S(math.Exp2(float64(precision)*8-1)-1)
	}
	return S(int64(xUint64)) / S(math.Exp2(float64(precision)*8-1)-1)
}

func unsignedToFloat[S Size](precision int, xUint64 uint64) S {
	return S(xUint64)/S(math.Exp2(float64(precision)*8)-1)*2 - 1
}

func norm[S Size](x S) S {
	if x < -1 {
		return -1
	}
	if x > +1 {
		return +1
	}
	return x
}

// Buffer is a storage for audio data. You can think of it as a bytes.Buffer for audio samples.
type Buffer[S Size, P Point[S]] struct {
	f    Format[S, P]
	data []byte
	tmp  []byte
}

// NewBuffer creates a new empty Buffer which stores samples in the provided format.
func NewBuffer[S Size, P Point[S]](f Format[S, P]) *Buffer[S, P] {
	return &Buffer[S, P]{f: f, tmp: make([]byte, f.Width())}
}

// Format returns the format of the Buffer.
func (b *Buffer[S, P]) Format() Format[S, P] {
	return b.f
}

// Len returns the number of samples currently in the Buffer.
func (b *Buffer[S, P]) Len() int {
	return len(b.data) / b.f.Width()
}

// Pop removes n samples from the beginning of the Buffer.
//
// Existing Streamers are not affected.
func (b *Buffer[S, P]) Pop(n int) {
	b.data = b.data[n*b.f.Width():]
}

// Append adds all audio data from the given Streamer to the end of the Buffer.
//
// The Streamer will be drained when this method finishes.
func (b *Buffer[S, P]) Append(s Streamer[S, P]) {
	var samples [512]P
	for {
		n, ok := s.Stream(samples[:])
		if !ok {
			break
		}
		for _, sample := range samples[:n] {
			b.f.EncodeSigned(b.tmp, sample)
			b.data = append(b.data, b.tmp...)
		}
	}
}

// Streamer returns a StreamSeeker which streams samples in the given interval (including from,
// excluding to). If from<0 or to>b.Len() or to<from, this method panics.
//
// When using multiple goroutines, synchronization of Streamers with the Buffer is not required,
// as Buffer is persistent (but efficient and garbage collected).
func (b *Buffer[S, P]) Streamer(from, to int) StreamSeeker[S, P] {
	return &bufferStreamer[S, P]{
		f:    b.f,
		data: b.data[from*b.f.Width() : to*b.f.Width()],
		pos:  0,
	}
}

type bufferStreamer[S Size, P Point[S]] struct {
	f    Format[S, P]
	data []byte
	pos  int
}

func (bs *bufferStreamer[S, P]) Stream(samples []P) (n int, ok bool) {
	if bs.pos >= len(bs.data) {
		return 0, false
	}
	for i := range samples {
		if bs.pos >= len(bs.data) {
			break
		}
		sample, advance := bs.f.DecodeSigned(bs.data[bs.pos:])
		samples[i] = sample
		bs.pos += advance
		n++
	}
	return n, true
}

func (bs *bufferStreamer[S, P]) Err() error {
	return nil
}

func (bs *bufferStreamer[S, P]) Len() int {
	return len(bs.data) / bs.f.Width()
}

func (bs *bufferStreamer[S, P]) Position() int {
	return bs.pos / bs.f.Width()
}

func (bs *bufferStreamer[S, P]) Seek(p int) error {
	if p < 0 || bs.Len() < p {
		return fmt.Errorf("buffer: seek position %v out of range [%v, %v]", p, 0, bs.Len())
	}
	bs.pos = p * bs.f.Width()
	return nil
}
