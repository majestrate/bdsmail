package model

import (
	"github.com/majestrate/bdsmail/lib/maildir"
)

// user credential
type User struct {
	// email address of user
	Email string `xorm:"pk"`
	// login credential, if empty string login is not allowed
	Login string
	// path to maildir
	MailDirPath string
}

// check if user's login is correct given password
func (u *User) CheckLogin(passwd string) (ok bool) {
	if len(u.Login) > 0 {
		ok = LoginCred(u.Login).Check(passwd)
	}
	return
}

// get user's maildir
func (u *User) MailDir() maildir.MailDir {
	return maildir.MailDir(u.MailDirPath)
}
