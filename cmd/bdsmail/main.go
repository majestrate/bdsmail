package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/majestrate/bdsmail/lib/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigchnl := make(chan os.Signal)
	log.SetLevel(log.InfoLevel)
	log.Info(server.Version())
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
		// start signal processer
		go func(s *server.Server) {
			for {
				sig, ok := <-sigchnl
				if ok {
					if sig == syscall.SIGHUP {
						// got sighup
						err := s.ReloadConfig()
						if err == nil {
							log.Info("Reloaded configuration")
						} else {
							log.Error("Failed to reload configuration ", err)
						}
					}
				} else {
					return
				}
			}
		}(s)
		signal.Notify(sigchnl, syscall.SIGHUP)
		if err == nil {
			log.Info("Starting Up Mail Server")
			s.Run()
		}
	}
	if err != nil {
		log.Error(err)
	}
}
