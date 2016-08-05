package server

import "fmt"

const VERS_MAJOR = 0
const VERS_MINOR = 1
const VERS_PATCH = 0
const SERV_NAME = "Brain Dead Simple Mail Server"


// return version string
func Version() string {
	return fmt.Sprintf("%s %d.%d.%d", SERV_NAME, VERS_MAJOR, VERS_MINOR, VERS_PATCH)
}
