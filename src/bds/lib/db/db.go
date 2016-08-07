package db

import (
	"bds/lib/maildir"
	"bds/lib/model"
)

// a callback that visits a user model safely
type UserVisitor func(*model.User) error

// a callback that initializes a new user model safely
type UserInitializer func(*model.User) error

// a callback that updates a user model, returns updated model or nil for "don't update"
type UserUpdater func(*model.User) *model.User

// db abstraction
type DB interface {
	// implements maildir.Getter
	maildir.Getter
	// implements pop3.UserAuthenticator
	CheckUserLogin(username, password string) (bool, error)
	// ensure that all migrations are done
	Ensure() error
	// visit every user and call a visitor
	VisitAllUsers(v UserVisitor) error
	// visit 1 user by email
	VisitUser(email string, v UserVisitor) error
	// create a new user, initialize values with i, visit after created with v
	CreateUser(i UserInitializer, v UserVisitor) error
	// ensure a user exists, intialize other members with i if it doesn't exist
	EnsureUser(name string, i UserInitializer) error
	// update a user that already exists, does nothing if it doesn't exist
	UpdateUser(name string, u UserUpdater) error

	// run db mainloop
	Run()
	// close access to database, all operations fail on this object after calling
	Close() error
}
