package db

import "github.com/majestrate/bdsmail/lib/model"

// a get user event, gets user by name
type userGetEvent struct {
	*dbEvent
	// name
	name string
	// the user fetched
	u *model.User
	// error
	err error
}

func (ev *userGetEvent) Error() error {
	return ev.err
}

func (ev *userGetEvent) Query() {
	var has bool
	u := new(model.User)
	has, ev.err = ev.X.engine.Id(ev.name).Get(u)
	if has {
		ev.u = u
	}
	return
}

type userInsertEvent struct {
	*dbEvent
	// user to insert
	u *model.User
	// any errors that occur
	err error
}

func (ev *userInsertEvent) Error() error {
	return ev.err
}

func (ev *userInsertEvent) Query() {
	_, ev.err = ev.X.engine.InsertOne(ev.u)
}

type userUpdateEvent struct {
	*dbEvent
	// user to update
	u *model.User
	// any errors that occur
	err error
}

func (ev *userUpdateEvent) Error() error {
	return ev.err
}

func (ev *userUpdateEvent) Query() {
	_, ev.err = ev.X.engine.Id(ev.u.Name).Update(ev.u)
}

// visit all users
type userVisitEvent struct {
	*dbEvent
	// visitor to call
	v UserVisitor
	// any error that occurs
	err error
}

func (ev *userVisitEvent) Error() error {
	return ev.err
}

func (ev *userVisitEvent) Query() {
	// for each every user ...
	ev.err = ev.X.engine.Where("name != ?", "").Iterate(new(model.User), func(_ int, u interface{}) error {
		// call visitor
		return ev.v(u.(*model.User))
	})
	return
}
