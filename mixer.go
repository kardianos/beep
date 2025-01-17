package beep

// Mixer allows for dynamic mixing of arbitrary number of Streamers. Mixer automatically removes
// drained Streamers. Mixer's stream never drains, when empty, Mixer streams silence.
type Mixer[S Size, P Point[S]] struct {
	streamers []Streamer[S, P]
}

// Len returns the number of Streamers currently playing in the Mixer.
func (m *Mixer[S, P]) Len() int {
	return len(m.streamers)
}

// Add adds Streamers to the Mixer.
func (m *Mixer[S, P]) Add(s ...Streamer[S, P]) {
	m.streamers = append(m.streamers, s...)
}

// Clear removes all Streamers from the mixer.
func (m *Mixer[S, P]) Clear() {
	m.streamers = m.streamers[:0]
}

// Stream streams all Streamers currently in the Mixer mixed together. This method always returns
// len(samples), true. If there are no Streamers available, this methods streams silence.
func (m *Mixer[S, P]) Stream(samples []P) (n int, ok bool) {
	var tmp [512]P
	var cP P
	ct := cP.Count()

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

		for si := 0; si < len(m.streamers); si++ {
			// mix the stream
			sn, sok := m.streamers[si].Stream(tmp[:toStream])
			for i := range tmp[:sn] {
				for j := 0; j < ct; j++ {
					samples[i].Add(j, tmp[i].Get(j))
				}
			}
			if !sok {
				// remove drained streamer
				sj := len(m.streamers) - 1
				m.streamers[si], m.streamers[sj] = m.streamers[sj], m.streamers[si]
				m.streamers = m.streamers[:sj]
				si--
			}
		}

		samples = samples[toStream:]
		n += toStream
	}

	return n, true
}

// Err always returns nil for Mixer.
//
// There are two reasons. The first one is that erroring Streamers are immediately drained and
// removed from the Mixer. The second one is that one Streamer shouldn't break the whole Mixer and
// you should handle the errors right where they can happen.
func (m *Mixer[S, P]) Err() error {
	return nil
}
