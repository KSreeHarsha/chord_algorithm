package triblab

type ChordMiniSnapshot struct {
	prev_prev		string
	prev			string
	me				string
	next			string
	next_next		string
	next_next_next	string
}

func (self *ChordMiniSnapshot) ofSizeOne() bool {
	return (self.prev == EMPTY_STRING)
}

func (self *ChordMiniSnapshot) ofSizeTwo() bool {
	return (self.prev == self.next)
}

func (self *ChordMiniSnapshot) ofSizeThree() bool {
	return (self.prev == self.prev_prev)
}

func (self *ChordMiniSnapshot) smallerThanFour() bool {
	return self.ofSizeOne() || self.ofSizeTwo() || self.ofSizeThree()
}

func CreateMiniChord(node string, chord *Chord) (ChordMiniSnapshot, error) {

	var c ChordMiniSnapshot
	var err error

	c.me = node

	// Prev
	c.prev, err = chord.Prev_node_ip(node)
	if err != nil { return c, err}

	// Next
	c.next, err = chord.Succ_node_ip(node)
	if err != nil { return c, err}

	// Try PrevPrev
	if c.prev != EMPTY_STRING {
		c.prev_prev, err = chord.Prev_node_ip(c.prev)
		if err != nil { return c, err}
	}

	// Try NextNext
	if c.next != EMPTY_STRING {
		c.next_next, err = chord.Succ_node_ip(c.next)
		if err != nil { return c, err}
	}

	// Try NextNextNext
	if c.next_next != EMPTY_STRING {
		c.next_next_next, err = chord.Succ_node_ip(c.next_next)
		if err != nil { return c, err}
	}

	return c, nil
}
