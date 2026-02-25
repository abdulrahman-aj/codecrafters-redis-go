package util

import "log"

func Assert(cond bool, msg string) {
	if !cond {
		log.Fatal(msg)
	}
}

func Fatal(msg string) {
	log.Fatal(msg)
}

func FatalOnErr(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func Retry(f func() error, n int) error {
	var err error
	for range n {
		err = f()
		if err == nil {
			return nil
		}
	}
	return err
}
