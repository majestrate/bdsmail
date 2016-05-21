package maildir

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"time"
)

// maildir mailbox protocol
type MailDir string

func (d MailDir) String() (str string) {
	str = string(d)
	return
}

// ensure the maildir is well formed
func (d MailDir) Ensure() (err error) {
	dir := d.String()
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		// create main directory
		err = os.Mkdir(dir, 0750)
	}
	if err == nil {
		// create subdirs
		for _, subdir := range []string{"new", "cur", "tmp"} {
			subdir = filepath.Join(d.String(), subdir)
			_, err = os.Stat(subdir)
			if os.IsNotExist(err) {
				// create non existant subdir
				err = os.Mkdir(subdir, 0750)
			}
		}
	}
	return
}

// get a string of the current filename to use
func (d MailDir) File() (fname string) {
	hostname, err := os.Hostname()
	if err == nil {
		fname = fmt.Sprintf("%d.%d.%s", time.Now().Unix(), os.Getpid(), hostname)
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
	f = filepath.Join(d.String(), "tmp", fname)
	return
}

func (d MailDir) NewFile() (fname string) {
	fname = d.New(d.File())
	return
}

func (d MailDir) New(fname string) (f string) {
	f = filepath.Join(d.String(), "new", fname)
	return
}

// deliver mail to this maildir
func (d MailDir) Deliver(body io.Reader) (err error) {
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
		err = os.Chdir(d.String())
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
				f, err = os.OpenFile(d.Temp(fname), os.O_CREATE|os.O_WRONLY, 0640)
				if err == nil {
					// write body
					_, err = io.Copy(f, body)
					f.Close()
					if err == nil {
						// now symlink
						err = os.Symlink(filepath.Join("..", "tmp", fname), filepath.Join("new", fname))
						// if err is nil it's delivered
					}
				}
			}
		}
	}
	return
}
