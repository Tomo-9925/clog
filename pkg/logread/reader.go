package logread

import (
	"bufio"
	"os"
	"time"

	"github.com/xapima/conps/pkg/util"
)

type Reader struct {
	offset   int64
	modTime  time.Time
	filePath string
	outCh    chan string
	errCh    chan error
}

func NewReader(outCh chan string, errCh chan error) *Reader {
	r := Reader{outCh: outCh, errCh: errCh}
	return &r
}

func (r *Reader) open(path string) {
	r.filePath = path
	r.offset = 0
	r.modTime = time.Unix(0, 0)
}

func (r *Reader) Read(path string) {
	defer close(r.outCh)
	defer close(r.errCh)

	r.open(path)
	for {
		if ok, err := r.isMod(); err != nil {
			r.errCh <- util.ErrorWrapFunc(err)
			return
		} else if !ok {
			continue
		}
		if err := r.read(); err != nil {
			r.errCh <- util.ErrorWrapFunc(err)
			return
		}
	}
}
func (r *Reader) read() error {
	if err := r.setModTime(); err != nil {
		return util.ErrorWrapFunc(err)
	}

	f, err := os.Open(r.filePath)
	if err != nil {
		return util.ErrorWrapFunc(err)
	}
	if _, err := f.Seek(r.offset, 0); err != nil {
		return util.ErrorWrapFunc(err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		r.outCh <- scanner.Text()
	}

	// set EOF offset
	ret, err := f.Seek(0, 2)
	if err != nil {
		return util.ErrorWrapFunc(err)
	}
	r.offset = ret

	return nil
}

func (r *Reader) setModTime() error {
	info, err := os.Stat(r.filePath)
	if err != nil {
		return util.ErrorWrapFunc(err)
	}
	r.modTime = info.ModTime()
	return nil
}

func (r *Reader) isMod() (bool, error) {
	info, err := os.Stat(r.filePath)
	if err != nil {
		return false, util.ErrorWrapFunc(err)
	}
	if r.modTime != info.ModTime() {
		return true, nil
	}
	return false, nil
}
