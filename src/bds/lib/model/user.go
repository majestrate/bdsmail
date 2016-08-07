package model

import (
	"bds/lib/maildir"
)

// mail user info
type User struct {
	// name of this user, aka the name part of name@ourb32address.b32.i2p
	Name string `xorm:"pk"`
	// login credential, if empty string login is not allowed
	Login string `xorm:"login"`
	// path to maildir
	MailDirPath string `xorm:"maildir"`
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

// ensure all resources for this user exist
// returns an error if one occurs while setting up any resources
func (u *User) Ensure() (err error) {
	err = u.MailDir().Ensure()
	return
}
