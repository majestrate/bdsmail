package main

import (
	"fmt"
	"os"
	"time"
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
	
	fmt.Printf("-- auto generated config made %s\n", time.Now().Format(time.ANSIC))
	fmt.Printf(`bind = "%s"`, bind)
	fmt.Printf("\n")
	fmt.Printf(`domain = "%s"`, domain)
	fmt.Printf("\n")
	fmt.Printf(`maildir = "%s"`, maildir)
	fmt.Printf("\n")
	for _, funcname := range []string {"whitelist", "blacklist", "checkspam"} {
		fmt.Printf("\n\nfunction %s(addr, recip, sender, body)\nreturn 0\nend\n\n", funcname)
	}
}
