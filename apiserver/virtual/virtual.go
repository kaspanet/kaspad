package virtual

var virtualBlueScore uint64

// SetBlueScore sets the blue score of the virtual
// to the given number.
func SetBlueScore(blueScore uint64) {
	virtualBlueScore = blueScore
}

// BlueScore returns the virtual blue score.
func BlueScore() uint64 {
	return virtualBlueScore
}
