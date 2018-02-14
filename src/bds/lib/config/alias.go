package config

import "bds/lib/config/parser"

type AliasConfig struct {
	fname string
}

func (c *AliasConfig) Load(fname string) (err error) {
	c.fname = fname
	return
}

func (c *AliasConfig) MX(hostname string) (addr string, ok bool) {
	conf, _ := parser.Read(c.fname)
	if conf != nil {
		s, _ := conf.Section(hostname)
		if s != nil {
			addr = s.ValueOf("mx")
			ok = addr != ""
		}
	}
	return
}
