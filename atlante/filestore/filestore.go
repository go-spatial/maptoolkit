package filestore

import (
	"io"
	"log"
	"net/url"
	"sync"
	"time"
)

const (
	// ConfigKeyName is the config key name for the name of a filestore entry
	ConfigKeyName = "name"
	// ConfigKeyType is the config key type for the type of a filestore entry
	ConfigKeyType = "type"
)

type timeout interface {
	Timeout() bool
}

// FileWriter returns a writer object
type FileWriter interface {
	// Writer should return an io.Writer that can be used to write the file to.
	// If the file should not be written to the filestore, return nil for
	// the io.WriteCloser.
	Writer(filepath string, isIntermediate bool) (io.WriteCloser, error)
}

// Provider returns a filestore that can be used to store files.
type Provider interface {
	// Writer provides a file writer that can be used to write file into the store.
	FileWriter(group string) (FileWriter, error)
}

type Exister interface {
	//Exists returns weather the fpath exists, and is not a directory.
	Exists(fpath string) bool
}

// URLInfo encodes the url that can be used to get the file and the LastModified time of that url.
type URLInfo struct {
	URL          *url.URL
	LastModified *time.Time
}

// String returns the string reprenstation of the URL
func (urlinfo URLInfo) String() string {
	if urlinfo.URL == nil {
		return ""
	}
	return urlinfo.URL.String()
}

// TimeString returns the LastModified time formated according to RFC3339
func (urlinfo URLInfo) TimeString() string {
	if urlinfo.LastModified == nil || urlinfo.LastModified.IsZero() {
		return ""
	}
	return urlinfo.LastModified.Format(time.RFC3339)
}

// Pather returns a url to the given file, the filestore supports external urls.
// If the file does not exist return nil for the url, and a ErrPath. This can
// be used for timeout as well. If the filestore does not support PathURLs
// (i.e. because of configuration) then return nil for the url and a ErrUnsupportedOperation
type Pather interface {
	PathURL(group string, filepath string, isIntermediate bool) (URLInfo, error)
}

// globalWaitGroupPipe is used by pipe to keep the process running
// till all the piped writes have had a chance to close and finish
// writing.
var globalWaitGroupPipe sync.WaitGroup

// Pipe creates a pipe that can be use to connect something that requires a io.Reader
func Pipe(typ, name string, fn func(r io.Reader) error) io.WriteCloser {
	r, w := io.Pipe()
	globalWaitGroupPipe.Add(1)
	go func() {
		// TODO(gdey): Need to figure something better for handeling errors here.
		// one idea is to use the CloseWithError() method, to pass the error to
		// the write side of the pipe.
		err := fn(r)
		if err != nil {
			log.Printf("error putting to %v (%v): %v", name, typ, err)
		}
		globalWaitGroupPipe.Done()
	}()
	return w
}

func cleanup() {
	globalWaitGroupPipe.Wait()
}
