package config

import "bds/lib/config/parser"

type AliasConfig struct {
	conf *parser.Configuration
}

func (c *AliasConfig) Load(fname string) (err error) {
	c.conf, err = parser.Read(fname)
	return
}

func (c *AliasConfig) MX(hostname string) (addr string, ok bool) {
	if c.conf != nil {
		s, _ := c.conf.Section(hostname)
		if s != nil {
			addr = s.ValueOf("mx")
			ok = addr != ""
		}
	}
	return
}
