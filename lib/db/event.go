package db

type dbQuery interface {
	// do the db query for this event
	Query()
	// get any errors that happened
	Error() error
	// wait for this query to complete
	Wait()
	// called when query is done
	Done()
}

type dbEvent struct {
	X *xormDB
	// completetion channel
	chnl chan bool
}

func (ev *dbEvent) Wait() {
	<-ev.chnl
	close(ev.chnl)
}

func (ev *dbEvent) Done() {
	ev.chnl <- true

}
