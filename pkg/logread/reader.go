package logread

import (
	"bytes"
	"io"
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
	defer f.Close()

	if _, err := f.Seek(r.offset, 0); err != nil {
		return util.ErrorWrapFunc(err)
	}
	// count := 0
	var ret int64
	for {

		buf := make([]byte, 1024*8)
		n, readErr := f.Read(buf)
		if readErr == io.EOF {
			break
		} else if readErr != nil {
			return util.ErrorWrapFunc(err)
		}
		if n == 0 {
			break
		}

		bufParts := bytes.Split(buf, []byte("\n"))
		ldx := len(bufParts) - 1
		// logrus.Debug("ldx:", ldx)
		// bufParts[ldx]は bufPartsの末尾が\nなら空、そうでなければ欠損データ

		// logrus.Debug("RAW_buf:\n", buf)
		// logrus.Debug("RAW:\n", string(buf))
		for i := 0; i < ldx; i++ {
			// if len(bufParts[i]) == 0 {
			// 	logrus.Debugf("break!! i:%d,", i)
			// 	break
			// }
			r.outCh <- string(bufParts[i])
		}

		if len(bufParts[ldx]) != 0 && bufParts[ldx][len(bufParts[ldx])-1] == 0 {
			for i := 0; i < len(bufParts[ldx]); i++ {
				if bufParts[ldx][i] == 0 {
					ret = int64(i)
					break
				}
			}
			// break
		} else {
			ret = int64(len(bufParts[ldx]))
		}
		// r.offset -= ret
		r.offset, err = f.Seek(-ret, 1)
		if err != nil {
			return util.ErrorWrapFunc(err)
		}
		// logrus.Debug("len last Part:", len(bufParts[ldx]))
		// if count == 100 {
		// 	panic("stop")
		// }
		// count++
		// if readErr == io.EOF {
		// 	logrus.Debug("EOF")
		// 	break
		// }
	}
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
