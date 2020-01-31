package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/xapima/clog/pkg/logread"
)

func main() {
	outCh := make(chan string)
	errCh := make(chan error)

	readPath := "./sample.txt"
	r, err := logread.NewReader(outCh, errCh)
	if err != nil {
		logrus.Error(err)
	}
	go r.Read(readPath)

L:
	for {
		select {
		case text := <-outCh:
			fmt.Println(text)
		case err := <-errCh:
			logrus.Error(err)
			break L
		}
	}

}
