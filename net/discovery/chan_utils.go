package discovery

// helper that blocks reading chan until a value is availble, then drains all and returns last.
func getMostRecentBool(c chan bool) (bool, bool) {
	// block until a value is available
	v, cont := <-c
	draining := cont
	for draining {
		select {
		case v, cont = <-c:
		default:
			draining = false
		}
	}
	return v, cont
}

// helper that blocks reading chan until a value is availble, then drains all and returns last.
func getMostRecentString(c chan string) (string, bool) {
	// block until a value is available
	v, cont := <-c
	draining := cont
	for draining {
		select {
		case v, cont = <-c:
		default:
			draining = false
		}
	}
	return v, cont
}
