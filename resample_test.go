package beep_test

import (
	"testing"

	"github.com/faiface/beep"
)

func TestResample(t *testing.T) {
	t.Run("float64-Stereo", runTestResample[float64, beep.Stereo[float64]])
	t.Run("float32-Stereo", runTestResample[float32, beep.Stereo[float32]])
	t.Run("float64-Mono", runTestResample[float64, beep.Mono[float64]])
	t.Run("float32-Mono", runTestResample[float32, beep.Mono[float32]])
}
func runTestResample[S beep.Size, P beep.Point[S]](t *testing.T) {
	var check float64 = 0.123456789123456
	delta := S(0)
	var cv S = S(check)
	if float64(cv) != check {
		delta = S(0.01)
	}

	for _, numSamples := range []int{8, 100, 500, 1000, 50000} {
		for _, old := range []beep.SampleRate{100, 2000, 44100, 48000} {
			for _, new := range []beep.SampleRate{100, 2000, 44100, 48000} {
				if numSamples/int(old)*int(new) > 1e6 {
					continue // skip too expensive combinations
				}

				s, data := randomDataStreamer[S, P](numSamples)

				want := resampleCorrect[S, P](3, old, new, data)

				got := collect[S, P](beep.Resample[S, P](3, old, new, s))

				if !equal(t, want, got, delta) {
					t.Fatalf("Resample not working correctly")
				}
			}
		}
	}
}

func equal[S beep.Size, P beep.Point[S]](t *testing.T, aa, bb []P, allowDelta S) bool {
	if len(aa) != len(bb) {
		return false
	}
	var cp P
	ct := cp.Count()
	for i := range aa {
		a, b := aa[i], bb[i]
		for j := 0; j < ct; j++ {
			va := a.Get(j)
			vb := b.Get(j)
			if va == vb {
				continue
			}
			diff := va - vb
			if diff < 0 {
				diff = -diff
			}
			if diff < allowDelta {
				continue
			}
			t.Errorf("%f != %f (delta %f)", va, vb, allowDelta)
			return false
		}
	}
	return true
}

func resampleCorrect[S beep.Size, P beep.Point[S]](quality int, old, new beep.SampleRate, p []P) []P {
	ratio := float64(old) / float64(new)
	pts := make([]point[S], quality*2)
	var resampled []P
	for i := 0; ; i++ {
		j := S(i) * S(ratio)
		if int(j) >= len(p) {
			break
		}
		var sample P
		sl := sample.Slice()
		for c := range sl {
			for k := range pts {
				l := int(j) + k - len(pts)/2 + 1
				if l >= 0 && l < len(p) {
					pts[k] = point[S]{X: S(l), Y: p[l].Get(c)}
				} else {
					pts[k] = point[S]{X: S(l), Y: 0}
				}
			}
			y := lagrange[S](pts[:], j)
			sample = sample.Set(c, y).(P)
		}
		resampled = append(resampled, sample)
	}
	return resampled
}

func lagrange[S beep.Size](pts []point[S], x S) (y S) {
	y = 0.0
	for j := range pts {
		l := S(1.0)
		for m := range pts {
			if j == m {
				continue
			}
			l *= (x - pts[m].X) / (pts[j].X - pts[m].X)
		}
		y += pts[j].Y * l
	}
	return y
}

type point[S beep.Size] struct {
	X, Y S
}
