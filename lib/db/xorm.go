package db

import (
	"github.com/go-xorm/xorm"
	"github.com/majestrate/bdsmail/lib/model"
	_ "github.com/mattn/go-sqlite3"
)

// create database driver
func NewDB(dburl string) (db DB, err error) {
	var eng *xorm.Engine
	eng, err = xorm.NewEngine("sqlite3", dburl)
	if err == nil {
		db = &xormDB{
			engine: eng,
		}
	}
	return
}

type xormDB struct {
	engine *xorm.Engine
}

func (x *xormDB) Ensure() (err error) {
	err = x.engine.Sync(new(model.User))
	return
}

func (x *xormDB) GetUser(email string) (u *model.User, err error) {
	var has bool
	u = new(model.User)
	u.Email = email
	has, err = x.engine.Get(u)
	if !has {
		u = nil
	}
	return
}

func (x *xormDB) Close() (err error) {
	if x.engine != nil {
		err = x.engine.Close()
	}
	x.engine = nil
	return
}
