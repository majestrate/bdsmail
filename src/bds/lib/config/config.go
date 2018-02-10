package config

import "bds/lib/config/parser"

type Config struct {
	opts map[string]string
}

func (c *Config) Get(name string) (val string, ok bool) {
	val, ok = c.opts[name]
	return
}

func (c *Config) Load(fname string) (err error) {
	var conf *parser.Configuration
	conf, err = parser.Read(fname)
	if err == nil {
		s, _ := conf.Section("")
		if s != nil {
			c.opts = s.Options()
		}
	}
	return
}
