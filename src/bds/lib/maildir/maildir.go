package maildir

import (
	"bds/lib/mailstore"
	"crypto/rand"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"time"
)

// maildir mailbox protocol
type MailDir string

// get absolute filepath for this maildir
func (d MailDir) Filepath() (str string) {
	str, _ = filepath.Abs(string(d))
	return
}

// ensure the maildir is well formed
func (d MailDir) Ensure() (err error) {
	dir := d.Filepath()
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		// create main directory
		err = os.Mkdir(dir, 0700)
	}
	if err == nil {
		// create subdirs
		for _, subdir := range []string{"new", "cur", "tmp"} {
			subdir = filepath.Join(dir, subdir)
			_, err = os.Stat(subdir)
			if os.IsNotExist(err) {
				// create non existant subdir
				err = os.Mkdir(subdir, 0700)
			}
		}
	}
	return
}

// get a string of the current filename to use
func (d MailDir) File() (fname string) {
	hostname, err := os.Hostname()
	if err == nil {
		b := make([]byte, 8)
		io.ReadFull(rand.Reader, b)
		fname = fmt.Sprintf("%x%d%d.%s", b, time.Now().Unix(), os.Getpid(), hostname)
	} else {
		log.Fatal("hostname() call failed", err)
	}
	return
}

func (d MailDir) TempFile() (fname string) {
	fname = d.Temp(d.File())
	return
}

func (d MailDir) Temp(fname string) (f string) {
	f = filepath.Join(d.Filepath(), "tmp", fname)
	return
}

func (d MailDir) NewFile() (fname string) {
	fname = d.New(d.File())
	return
}

func (d MailDir) New(fname string) (f string) {
	f = filepath.Join(d.Filepath(), "new", fname)
	return
}

func (d MailDir) Cur(fname string) (f string) {
	f = filepath.Join(d.Filepath(), "cur", fname)
	return
}

// deliver mail to this maildir
// return messsage that was delivered
func (d MailDir) Deliver(body io.Reader) (msg mailstore.Message, err error) {
	var oldwd string
	oldwd, err = os.Getwd()
	if err == nil {
		// no error getting working directory, let's begin

		// when done chdir to previous directory
		defer func() {
			err := os.Chdir(oldwd)
			if err != nil {
				log.Fatal("chdir failed", err)
			}
		}()
		// chdir to maildir
		err = os.Chdir(d.Filepath())
		if err == nil {
			fname := d.File()
			for {
				_, err = os.Stat(d.Temp(fname))
				if os.IsNotExist(err) {
					break
				}
				time.Sleep(time.Second * 2)
				fname = d.File()
			}
			// set err to nil
			err = nil
			var f *os.File
			// create tmp file
			f, err = os.Create(d.Temp(fname))
			if err == nil {
				// success creation
				err = f.Close()
			}
			// try writing file
			if err == nil {
				f, err = os.OpenFile(d.Temp(fname), os.O_CREATE|os.O_WRONLY, 0600)
				if err == nil {
					// write body
					_, err = io.Copy(f, body)
					f.Close()
					if err == nil {
						fn := d.New(fname)
						err = os.Rename(d.Temp(fname), fn)
						if err == nil {
							// delivered
							msg = Message(fn)
						}
					}
				}
			}
		}
	}
	return
}

// list messages in subdirectory
func (d MailDir) listDir(sd string) (msgs []Message, err error) {
	var f *os.File
	fp := filepath.Join(d.Filepath(), sd)
	f, err = os.Open(fp)
	if err == nil {
		defer f.Close()
		var files []string
		files, err = f.Readdirnames(0)
		for _, mf := range files {
			mf = filepath.Join(fp, mf)
			msgs = append(msgs, Message(mf))
		}
	}
	return
}

// list new messages in this maildir
func (d MailDir) ListNew() (msgs []mailstore.Message, err error) {
	var m []Message
	m, err = d.listDir("new")
	if err == nil {
		for _, msg := range m {
			msgs = append(msgs, msg)
		}
	}
	return
}

// list currently held messages in this maildir
func (d MailDir) ListCur() (msgs []Message, err error) {
	msgs, err = d.listDir("cur")
	return
}

// process new message and move it to the cur directory
func (d MailDir) ProcessNew(msg Message, flags ...Flag) (m Message, err error) {
	// find message
	fname := d.New(msg.Filename())
	_, err = os.Stat(fname)
	if err == nil {
		var newname string
		// message exists and is accessable
		if len(flags) > 0 {
			var fl string
			for _, f := range flags {
				fl += f.String()
			}
			// set flags
			newname = d.Cur(fmt.Sprintf("%s:2,%s", msg.Name(), fl))
		} else {
			// don't touch flags
			newname = d.Cur(msg.Filename())
		}
		err = os.Rename(fname, newname)
		if err == nil {
			m = Message(newname)
		}
	}
	return
}

// process message in cur and change its flags if specified
func (d MailDir) ProcessCur(msg Message, flags ...Flag) (err error) {
	fname := d.Cur(msg.Filename())
	_, err = os.Stat(fname)
	if err == nil {
		// message exists and is accessable
		if len(flags) > 0 {
			var fl string
			for _, f := range flags {
				fl += f.String()
			}
			// set message flags
			err = os.Rename(fname, d.Cur(fmt.Sprintf("%s:2,%s", msg.Name(), fl)))
		} else {
			// don't touch the message's flags if non are provided
		}
	}
	return
}

// return true if this message is in cur directory
func (d MailDir) IsCur(msg Message) (is bool, err error) {
	_, err = os.Stat(d.Cur(msg.Filepath()))
	if os.IsNotExist(err) {
		err = nil
	} else {
		is = true
	}
	return
}

// return true if this message is in cur directory
func (d MailDir) IsNew(msg Message) (is bool, err error) {
	_, err = os.Stat(d.New(msg.Filepath()))
	if os.IsNotExist(err) {
		err = nil
	} else {
		is = true
	}
	return
}

// open message in cur directory
func (d MailDir) OpenMessage(msg Message) (f *os.File, err error) {
	f, err = os.Open(d.Cur(msg.Filepath()))
	return
}
