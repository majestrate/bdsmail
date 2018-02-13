package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// bds mail configuration generator

func main() {
	domain := "localhost"
	bindmail := "127.0.0.1:2525"
	bindweb := "127.0.0.1:8888"
	maildir := "mail"
	asset_dir := filepath.Join(".", "contrib", "assets", "web")
	i2paddr := "127.0.0.1:7656"
	keyfile := "bdsmail-privkey.dat"

	if len(os.Args) > 1 {
		domain = os.Args[1]
	}
	if len(os.Args) > 2 {
		maildir = os.Args[2]
	}
	fmt.Println("[maild]")
	fmt.Printf(`i2paddr = %s`, i2paddr)
	fmt.Printf("\n")
	fmt.Printf(`i2pkeyfile = %s`, keyfile)
	fmt.Printf("\n")
	fmt.Printf(`bindmail = %s`, bindmail)
	fmt.Printf("\n")
	fmt.Printf(`bindweb = %s`, bindweb)
	fmt.Printf("\n")
	fmt.Printf(`domain = %s`, domain)
	fmt.Printf("\n")
	fmt.Printf(`maildir = %s`, maildir)
	fmt.Printf("\n")
	fmt.Printf(`database = %s.sqlite`, domain)
	fmt.Printf("\n")
	fmt.Printf(`assets = %s`, asset_dir)
	fmt.Printf("\n")
}
