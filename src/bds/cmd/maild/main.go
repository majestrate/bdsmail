package main

import (
	"bds/lib/server"
	log "github.com/Sirupsen/logrus"
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
	// load config
	err := s.LoadConfig(cfg_fname)
	if err == nil {
		// bind server
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
					} else if sig == syscall.SIGTERM || sig == syscall.SIGINT {
						log.Info("Stopping Server")
						s.Stop()
					}
				} else {
					return
				}
			}
		}(s)
		if err == nil {
			// set signal handler
			signal.Notify(sigchnl, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

			log.Info("Starting Up Mail Server")
			// run server
			s.Run()
			log.Info("Mail Server done")
			os.Exit(0)
		}
	}
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
