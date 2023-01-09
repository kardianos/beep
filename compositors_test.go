package beep_test

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/faiface/beep"
)

// randomDataStreamer generates random samples of duration d and returns a StreamSeeker which streams
// them and the data itself.
func randomDataStreamer[S beep.Size, P beep.Point[S]](numSamples int) (s beep.StreamSeeker[S, P], data []P) {
	data = make([]P, numSamples)
	for i := range data {
		d := data[i]
		for c := range d.Slice() {
			data[i] = d.Set(c, S(rand.Float64()*2-1)).(P)
		}
	}
	return &dataStreamer[S, P]{data, 0}, data
}

type dataStreamer[S beep.Size, P beep.Point[S]] struct {
	data []P
	pos  int
}

func (ds *dataStreamer[S, P]) Stream(samples []P) (n int, ok bool) {
	if ds.pos >= len(ds.data) {
		return 0, false
	}
	n = copy(samples, ds.data[ds.pos:])
	ds.pos += n
	return n, true
}

func (ds *dataStreamer[S, P]) Err() error {
	return nil
}

func (ds *dataStreamer[S, P]) Len() int {
	return len(ds.data)
}

func (ds *dataStreamer[S, P]) Position() int {
	return ds.pos
}

func (ds *dataStreamer[S, P]) Seek(p int) error {
	ds.pos = p
	return nil
}

// collect drains Streamer s and returns all of the samples it streamed.
func collect[S beep.Size, P beep.Point[S]](s beep.Streamer[S, P]) []P {
	var (
		result []P
		buf    [479]P
	)
	for {
		n, ok := s.Stream(buf[:])
		if !ok {
			return result
		}
		result = append(result, buf[:n]...)
	}
}

func TestTake(t *testing.T) {
	for i := 0; i < 7; i++ {
		total := rand.Intn(1e5) + 1e4
		s, data := randomDataStreamer[float64, beep.Stereo[float64]](total)
		take := rand.Intn(total)

		want := data[:take]
		got := collect(beep.Take[float64, beep.Stereo[float64]](take, s))

		if !reflect.DeepEqual(want, got) {
			t.Error("Take not working correctly")
		}
	}
}

func TestLoop(t *testing.T) {
	for i := 0; i < 7; i++ {
		for n := 0; n < 5; n++ {
			s, data := randomDataStreamer[float64, beep.Stereo[float64]](10)

			var want []beep.Stereo[float64]
			for j := 0; j < n; j++ {
				want = append(want, data...)
			}
			got := collect(beep.Loop(n, s))

			if !reflect.DeepEqual(want, got) {
				t.Error("Loop not working correctly")
			}
		}
	}
}

func TestSeq(t *testing.T) {
	var (
		n    = 7
		s    = make([]beep.Streamer[float64, beep.Stereo[float64]], n)
		data = make([][]beep.Stereo[float64], n)
	)
	for i := range s {
		s[i], data[i] = randomDataStreamer[float64, beep.Stereo[float64]](rand.Intn(1e5) + 1e4)
	}

	var want []beep.Stereo[float64]
	for _, d := range data {
		want = append(want, d...)
	}

	got := collect(beep.Seq(s...))

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Seq not working properly")
	}
}

func TestMix(t *testing.T) {
	t.Run("float64-Stereo", runTestMix[float64, beep.Stereo[float64]])
	t.Run("float32-Stereo", runTestMix[float32, beep.Stereo[float32]])
	t.Run("float64-Mono", runTestMix[float64, beep.Mono[float64]])
	t.Run("float32-Mono", runTestMix[float32, beep.Mono[float32]])
}
func runTestMix[S beep.Size, P beep.Point[S]](t *testing.T) {
	var (
		n    = 7
		s    = make([]beep.Streamer[S, P], n)
		data = make([][]P, n)
	)
	for i := range s {
		s[i], data[i] = randomDataStreamer[S, P](rand.Intn(1e5) + 1e4)
	}

	maxLen := 0
	for _, d := range data {
		if len(d) > maxLen {
			maxLen = len(d)
		}
	}

	want := make([]P, maxLen)
	var cp P
	ct := cp.Count()
	for _, dd := range data {
		for i := range dd {
			w, d := want[i], dd[i]
			for j := 0; j < ct; j++ {
				w.Add(j, d.Get(j))
			}
		}
	}

	got := collect(beep.Mix(s...))

	if !reflect.DeepEqual(want, got) {
		t.Error("Mix not working correctly")
	}
}

func TestDup(t *testing.T) {
	t.Run("float64-Stereo", runTestDup[float64, beep.Stereo[float64]])
	t.Run("float32-Stereo", runTestDup[float32, beep.Stereo[float32]])
	t.Run("float64-Mono", runTestDup[float64, beep.Mono[float64]])
	t.Run("float32-Mono", runTestDup[float32, beep.Mono[float32]])
}
func runTestDup[S beep.Size, P beep.Point[S]](t *testing.T) {
	for i := 0; i < 7; i++ {
		s, data := randomDataStreamer[S, P](rand.Intn(1e5) + 1e4)
		st, su := beep.Dup[S, P](s)

		var tData, uData []P
		for {
			buf := make([]P, rand.Intn(1e4))
			tn, tok := st.Stream(buf)
			tData = append(tData, buf[:tn]...)

			buf = make([]P, rand.Intn(1e4))
			un, uok := su.Stream(buf)
			uData = append(uData, buf[:un]...)

			if !tok && !uok {
				break
			}
		}

		if !reflect.DeepEqual(data, tData) || !reflect.DeepEqual(data, uData) {
			t.Error("Dup not working correctly")
		}
	}
}
