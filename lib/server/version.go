package server

import "fmt"

const VERS_MAJOR = 0
const VERS_MINOR = 1
const VERS_PATCH = 0

// return version string
func Version() string {
	return fmt.Sprintf("BrainDeadSimple Mail Server %d.%d.%d", VERS_MAJOR, VERS_MINOR, VERS_PATCH)
}
