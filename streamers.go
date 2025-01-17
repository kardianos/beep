package beep

// Silence returns a Streamer which streams num samples of silence. If num is negative, silence is
// streamed forever.
func Silence[S Size, P Point[S]](num int) Streamer[S, P] {
	return StreamerFunc[S, P](func(samples []P) (n int, ok bool) {
		if num == 0 {
			return 0, false
		}
		if 0 < num && num < len(samples) {
			samples = samples[:num]
		}
		for i := range samples {
			var p P
			samples[i] = p
		}
		if num > 0 {
			num -= len(samples)
		}
		return len(samples), true
	})
}

// Callback returns a Streamer, which does not stream any samples, but instead calls f the first
// time its Stream method is called. The speaker is locked while f is called.
func Callback[S Size, P Point[S]](f func()) Streamer[S, P] {
	return StreamerFunc[S, P](func(samples []P) (n int, ok bool) {
		if f != nil {
			f()
			f = nil
		}
		return 0, false
	})
}

// Iterate returns a Streamer which successively streams Streamers obtains by calling the provided g
// function. The streaming stops when g returns nil.
//
// Iterate does not propagate errors from the generated Streamers.
func Iterate[S Size, P Point[S]](g func() Streamer[S, P]) Streamer[S, P] {
	var (
		s     Streamer[S, P]
		first = true
	)
	return StreamerFunc[S, P](func(samples []P) (n int, ok bool) {
		if first {
			s = g()
			first = false
		}
		if s == nil {
			return 0, false
		}
		for len(samples) > 0 {
			if s == nil {
				break
			}
			sn, sok := s.Stream(samples)
			if !sok {
				s = g()
			}
			samples = samples[sn:]
			n += sn
		}
		return n, true
	})
}
