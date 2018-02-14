package config

import (
	"bds/lib/config/parser"
	log "github.com/Sirupsen/logrus"
)

type Config struct {
	opts    map[string]string
	Aliases AliasConfig
}

func (c *Config) Get(name string) (val string, ok bool) {
	val, ok = c.opts[name]
	return
}

func (c *Config) Load(fname string) (err error) {
	var conf *parser.Configuration
	conf, err = parser.Read(fname)
	if err == nil {
		s, _ := conf.Section("maild")
		if s != nil {
			c.opts = s.Options()
			a := s.ValueOf("aliases")
			if a != "" {
				log.Infof("Loading Aliases %s", a)
				err = c.Aliases.Load(a)
			}
		}
	}
	return
}
