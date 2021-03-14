package main

import (
	"github.com/majestrate/bdsmail/lib/config"
	"github.com/majestrate/bdsmail/lib/db"
	"github.com/majestrate/bdsmail/lib/maildir"
	"github.com/majestrate/bdsmail/lib/model"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func main() {

	if len(os.Args) < 4 {
		log.Errorf("Usage: %s config.ini username maildirpath [password]", os.Args[0])
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
	if _, err = os.Stat(cfg_fname); err != nil {
		log.Errorf("failed to load config: %s", err.Error())
		return
	}
	err = conf.Load(cfg_fname)
	if err != nil {
		log.Errorf("Failed to parse %s, %s", cfg_fname, err.Error())
	}

	s, _ := conf.Get("database")
	if s == "" {
		log.Error("no database provided")
		return
	}
	db, err := db.NewDB(s)
	if err != nil {
		log.Errorf("failed to open db: %s", err.Error())
		return
	}
	log.Infof("opened %s", s)
	go db.Run()
	err = db.EnsureUser(user, func(u *model.User) error {
		log.Infof("creating user: %s", u.Name)
		return nil
	})
	if err != nil {
		log.Errorf("error creating user: %s", err.Error())
	}

	err = db.UpdateUser(user, func(u *model.User) *model.User {
		u.MailDirPath = m
		log.Infof("setting %s maildir to %s", user, m)
		return u
	})

	if err != nil {
		log.Errorf("Failed to set maildir: %s", err.Error())
	}

	if len(passwd) > 0 {
		err = db.UpdateUser(user, func(u *model.User) *model.User {
			u.Login = string(model.NewLoginCred(passwd))
			log.Infof("upading %s password", user)
			return u
		})
	}
	db.Close()

	if err == nil {
		log.Info("OK")
	} else {
		log.Errorf("error: %s", err.Error())
	}
}
