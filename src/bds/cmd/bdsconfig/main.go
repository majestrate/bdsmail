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
	fmt.Println()
	fmt.Printf(`i2pkeyfile = %s`, keyfile)
	fmt.Println()
	fmt.Printf(`bindmail = %s`, bindmail)
	fmt.Println()
	fmt.Printf(`bindweb = %s`, bindweb)
	fmt.Println()
	fmt.Printf(`bindpop3 = 127.0.0.1:1110`)
	fmt.Println()
	fmt.Printf(`domain = %s`, domain)
	fmt.Println()
	fmt.Printf(`maildir = %s`, maildir)
	fmt.Println()
	fmt.Printf(`database = %s.sqlite`, domain)
	fmt.Println()
	fmt.Printf(`assets = %s`, asset_dir)
	fmt.Println()
}
