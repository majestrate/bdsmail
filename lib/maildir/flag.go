package maildir

// maildir flag
type Flag rune

func (f Flag) Rune() rune {
	return rune(f)
}

func (f Flag) String() string {
	return string(f)
}

const Passed = Flag('P')
const Replied = Flag('R')
const Seen = Flag('S')
const Trashed = Flag('T')
const Draft = Flag('D')
const Flagged = Flag('F')
