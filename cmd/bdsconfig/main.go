package main

import (
	"fmt"
	"os"
)

// bds mail configuration generator

func main() {
	domain := "localhost"
	bind := ":2525"
	maildir := "mail"

	if len(os.Args) > 1 {
		domain = os.Args[1]
	}
	if len(os.Args) > 2 {
		maildir = os.Args[2]
	}

	fmt.Fprintf("-- auto generated config made %s\n", time.Now().Format(time.ANSIC))
	fmt.Fprintf(`bind = ":%s"\n`, bind)
	fmt.Fprintf(`domain = "%s"\n`, domain)
	fmt.Fprintf(`maildir = "%s"\n`, maildir)
	for _, funcname := range []string{"whitelist", "blacklist", "checkspam"} {
		fmt.Fprintf(`\n\nfunction %s(addr, recip, sender, body)\nreturn 0\nend\n\n`, funcname)
	}
}
