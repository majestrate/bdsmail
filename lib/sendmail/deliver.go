package sendmail

type DeliverJob interface {
	// cancel delivery job
	Cancel()
	// wait for completion, return true if delivered otherwise false
	Wait() bool
	// run delivery
	Run()
}
