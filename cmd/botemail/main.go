package main

import (
	"github.com/majestrate/botemail/lib/server"
	log "github.com/Sirupsen/logrus"
	"os"
)


func main() {
	log.SetLevel(log.DebugLevel)
	var cfg_fname string
	if len(os.Args) == 1 {
		// no args
		log.Fatal("no config file specified")
	} else {
		cfg_fname = os.Args[1]
	}

	s := server.New()
	err := s.LoadConfig(cfg_fname)
	if err == nil {
		err = s.Bind()
		if err == nil {
			log.Info("Starting Up Mail Server")
			s.Run()
		}
	}
	if err != nil {
		log.Error(err)
	}
}
