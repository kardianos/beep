package beep

// Take returns a Streamer which streams at most num samples from s.
//
// The returned Streamer propagates s's errors through Err.
func Take[S Size, P Point[S]](num int, s Streamer[S, P]) Streamer[S, P] {
	return &take[S, P]{
		s:       s,
		remains: num,
	}
}

type take[S Size, P Point[S]] struct {
	s       Streamer[S, P]
	remains int
}

func (t *take[S, P]) Stream(samples []P) (n int, ok bool) {
	if t.remains <= 0 {
		return 0, false
	}
	toStream := t.remains
	if len(samples) < toStream {
		toStream = len(samples)
	}
	n, ok = t.s.Stream(samples[:toStream])
	t.remains -= n
	return n, ok
}

func (t *take[S, P]) Err() error {
	return t.s.Err()
}

// Loop takes a StreamSeeker and plays it count times. If count is negative, s is looped infinitely.
//
// The returned Streamer propagates s's errors.
func Loop[S Size, P Point[S]](count int, s StreamSeeker[S, P]) Streamer[S, P] {
	return &loop[S, P]{
		s:       s,
		remains: count,
	}
}

type loop[S Size, P Point[S]] struct {
	s       StreamSeeker[S, P]
	remains int
}

func (l *loop[S, P]) Stream(samples []P) (n int, ok bool) {
	if l.remains == 0 || l.s.Err() != nil {
		return 0, false
	}
	for len(samples) > 0 {
		sn, sok := l.s.Stream(samples)
		if !sok {
			if l.remains > 0 {
				l.remains--
			}
			if l.remains == 0 {
				break
			}
			err := l.s.Seek(0)
			if err != nil {
				return n, true
			}
			continue
		}
		samples = samples[sn:]
		n += sn
	}
	return n, true
}

func (l *loop[S, P]) Err() error {
	return l.s.Err()
}

// Seq takes zero or more Streamers and returns a Streamer which streams them one by one without pauses.
//
// Seq does not propagate errors from the Streamers.
func Seq[S Size, P Point[S]](s ...Streamer[S, P]) Streamer[S, P] {
	i := 0
	return StreamerFunc[S, P](func(samples []P) (n int, ok bool) {
		for i < len(s) && len(samples) > 0 {
			sn, sok := s[i].Stream(samples)
			samples = samples[sn:]
			n, ok = n+sn, ok || sok
			if !sok {
				i++
			}
		}
		return n, ok
	})
}

// Mix takes zero or more Streamers and returns a Streamer which streams them mixed together.
//
// Mix does not propagate errors from the Streamers.
func Mix[S Size, P Point[S]](s ...Streamer[S, P]) Streamer[S, P] {
	return StreamerFunc[S, P](func(samples []P) (n int, ok bool) {
		var tmp [512]P
		for len(samples) > 0 {
			toStream := len(tmp)
			if toStream > len(samples) {
				toStream = len(samples)
			}

			// clear the samples
			for i := range samples[:toStream] {
				var p P
				samples[i] = p
			}

			snMax := 0 // max number of streamed samples in this iteration
			for _, st := range s {
				// mix the stream
				sn, sok := st.Stream(tmp[:toStream])
				if sn > snMax {
					snMax = sn
				}
				ok = ok || sok

				for i := range tmp[:sn] {
					samples[i][0] += tmp[i][0]
					samples[i][1] += tmp[i][1]
				}
			}

			n += snMax
			if snMax < len(tmp) {
				break
			}
			samples = samples[snMax:]
		}

		return n, ok
	})
}

// Dup returns two Streamers which both stream the same data as the original s. The two Streamers
// can't be used concurrently without synchronization.
func Dup[S Size, P Point[S]](s Streamer[S, P]) (t, u Streamer[S, P]) {
	var tBuf, uBuf []P
	return &dup[S, P]{&tBuf, &uBuf, s}, &dup[S, P]{&uBuf, &tBuf, s}
}

type dup[S Size, P Point[S]] struct {
	myBuf, itsBuf *[]P
	s             Streamer[S, P]
}

func (d *dup[S, P]) Stream(samples []P) (n int, ok bool) {
	buf := *d.myBuf
	n = copy(samples, buf)
	ok = len(buf) > 0
	buf = buf[n:]
	samples = samples[n:]
	*d.myBuf = buf

	if len(samples) > 0 {
		sn, sok := d.s.Stream(samples)
		n += sn
		ok = ok || sok
		*d.itsBuf = append(*d.itsBuf, samples[:sn]...)
	}

	return n, ok
}

func (d *dup[S, P]) Err() error {
	return d.s.Err()
}
