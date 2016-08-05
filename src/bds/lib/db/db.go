package db

import (
	"bds/lib/model"
)

// db abstraction
type DB interface {
	// ensure that all migrations are done
	Ensure() error
	// get a user by email address
	// returns nil model if user does not exist
	GetUser(email string) (*model.User, error)
	// close access to database, all operations fail on this object after calling
	Close() error
}
