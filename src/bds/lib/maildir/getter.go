package maildir

// defines a way to get a maildir given a username
type Getter interface {
	// get a user's maildir
	GetMailDir(user string) (MailDir, error)
}

type mailDirGetter string

func (md mailDirGetter) GetMailDir(user string) (m MailDir, err error) {
	m = MailDir(md)
	return
}

// a maildir getter that always uses 1 directory
func AbsoluteGetter(path string) Getter {
	return mailDirGetter(path)
}
