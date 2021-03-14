package db

import (
	"github.com/majestrate/bdsmail/lib/mailstore"
	"github.com/majestrate/bdsmail/lib/model"
	log "github.com/sirupsen/logrus"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

type xormDB struct {
	engine      *xorm.Engine
	dbChnl      chan dbQuery
	closewaiter chan error
	runit       bool
}

// create database driver
func NewDB(dburl string) (db DB, err error) {
	var eng *xorm.Engine
	eng, err = xorm.NewEngine("sqlite3", dburl)
	if err == nil {
		db = &xormDB{
			engine:      eng,
			dbChnl:      make(chan dbQuery, 32),
			closewaiter: make(chan error),
		}
	}
	return
}

func (x *xormDB) Ensure() (err error) {
	// ensure underlying xorm engine
	err = x.engine.Sync(new(model.User))
	return
}

func (x *xormDB) CheckUserLogin(username, password string) (good bool, err error) {
	var u *model.User
	u, err = x.getUser(username)
	if err == nil && u != nil && u.Login != "" {
		good = model.LoginCred(u.Login).Check(password)
	}
	return
}

func (x *xormDB) EnsureUser(name string, i UserInitializer) (err error) {
	if _, err = x.getUser(name); err != nil {
		// already there
		log.Infof("already have user %s", name)
		return
	}
	u := &model.User{
		Name: name,
	}
	if i != nil {
		err = i(u)
	}
	if err == nil {
		// make sure the name is still set
		u.Name = name
		ev := &userInsertEvent{
			dbEvent: &dbEvent{
				X:    x,
				chnl: make(chan bool),
			},
			u: u,
		}
		x.fireEvent(ev)
		ev.Wait()
	}
	return
}

func (x *xormDB) UpdateUser(email string, up UserUpdater) (err error) {
	var u *model.User
	u, err = x.getUser(email)
	if err == nil && u != nil {
		u = up(u)
		if u != nil && err == nil {
			// commit
			ev := &userUpdateEvent{
				dbEvent: &dbEvent{
					X:    x,
					chnl: make(chan bool),
				},
				u: u,
			}
			// fire and collect
			if x.fireEvent(ev) {
				ev.Wait()
				err = ev.Error()
			}
		}
	}
	return
}

// get maildir for user given email
func (x *xormDB) FindStoreFor(email string) (st mailstore.Store, has bool) {
	u, _ := x.getUser(email)
	if u != nil {
		st = u.MailDir()
		has = true
	}
	return
}

func (x *xormDB) VisitUser(email string, v UserVisitor) (err error) {
	var u *model.User
	u, err = x.getUser(email)
	if u != nil && err == nil {
		err = v(u)
	}
	return
}

func (x *xormDB) VisitAllUsers(v UserVisitor) (err error) {
	ev := &userVisitEvent{
		dbEvent: &dbEvent{
			X:    x,
			chnl: make(chan bool),
		},
		v: v,
	}
	x.fireEvent(ev)
	ev.Wait()
	return
}

// fire a db event and return true if it was queued otherwise false
func (x *xormDB) fireEvent(ev dbQuery) bool {
	if x.dbChnl != nil {
		x.dbChnl <- ev
	}
	return x.dbChnl != nil
}

func (x *xormDB) CreateUser(i UserInitializer, v UserVisitor) (err error) {
	u := new(model.User)
	if i != nil {
		err = i(u)
	}
	if err == nil {
		// insert user
		ev := &userInsertEvent{
			dbEvent: &dbEvent{
				X:    x,
				chnl: make(chan bool),
			},
			u: u,
		}
		x.fireEvent(ev)
		ev.Wait()
	}

	if err == nil && v != nil {
		err = v(u)
	}
	return
}

// safely get 1 user by email address
func (x *xormDB) getUser(email string) (u *model.User, err error) {
	name := strings.Split(email, "@")[0]
	idx := strings.Index(name, "<")
	if idx != -1 {
		name = name[1+idx:]
	}
	ev := &userGetEvent{
		name: name,
		dbEvent: &dbEvent{
			X:    x,
			chnl: make(chan bool),
		},
	}
	// fire event
	if x.fireEvent(ev) {
		// wait for result
		ev.Wait()
		err = ev.Error()
		u = ev.u
	}
	return
}

func (x *xormDB) Run() {
	for x.dbChnl != nil {
		ev, ok := <-x.dbChnl
		if !ok {
			break
		}
		ev.Query()
		ev.Done()
	}
	// close
	x.closewaiter <- x.engine.Close()
	x.engine = nil
}

func (x *xormDB) Close() (err error) {
	c := x.dbChnl
	x.dbChnl = nil
	close(c)
	err = <-x.closewaiter
	close(x.closewaiter)
	return
}
