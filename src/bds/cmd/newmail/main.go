package main

import (
	"bds/lib/config"
	"bds/lib/db"
	"bds/lib/maildir"
	"bds/lib/model"
	log "github.com/Sirupsen/logrus"
	"os"
	"path/filepath"
)

func main() {

	if len(os.Args) < 4 {
		log.Errorf("Usage: %s config.lua username maildirpath [password]", os.Args[0])
		return
	}

	cfg_fname := os.Args[1]
	user := os.Args[2]
	m, _ := filepath.Abs(os.Args[3])
	passwd := ""
	if len(os.Args) == 5 {
		passwd = os.Args[4]
	}
	md := maildir.MailDir(m)
	err := md.Ensure()
	if err != nil {
		log.Errorf("failed to create maildir: %s", err.Error())
		return
	}
	conf := new(config.Config)
	conf.Load(cfg_fname)

	s, ok := conf.Get("database")
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
			if len(passwd) > 0 {
				u.Login = string(model.NewLoginCred(passwd))
			}
			return nil
		})
		if err != nil {
			log.Errorf("error creating user: %s", err.Error())
		}
		db.Close()
	}
}
