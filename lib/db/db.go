package db

// db abstraction
type DB interface {
	// ensure that all migrations are done
	Ensure() error
}
