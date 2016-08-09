package main

import (
	log "github.com/Sirupsen/logrus"
	"bds/lib/db"
	"bds/lib/lua"
	"bds/lib/maildir"
	"bds/lib/model"
	"os"
	"path/filepath"
)


func main() {

	if len(os.Args) != 4 {
		log.Errorf("Usage: %s config.lua username maildirpath", os.Args[0])
		return
	}
	
	cfg_fname := os.Args[1]
	user := os.Args[2]
	m, _ := filepath.Abs(os.Args[3])
	md := maildir.MailDir(m)
	err := md.Ensure()
	if err != nil {
		log.Errorf("failed to create maildir: %s", err.Error())
		return
	}
	l := lua.New()
	l.JIT()
	l.LoadFile(cfg_fname)
	
	s, ok := l.GetConfigOpt("database")
	if ok {
		db, err := db.NewDB(s)
		if err != nil {
			log.Errorf("failed to open db: %s", err.Error())
			return
		}
		go db.Run()
		err = db.EnsureUser(user, func(u *model.User) error {
			log.Infof("creating user: %s", u.Name)
			u.MailDirPath = m
			return nil
		})
		if err != nil {
			log.Errorf("error creating user: %s", err.Error())
		}
		db.Close()
	}
	l.Close()
}
