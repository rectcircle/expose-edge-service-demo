package helper

import "io"

func IORelay(a, b io.ReadWriter) error {
	errc := make(chan error, 1)
	go func() {
		_, err := io.Copy(a, b)
		errc <- err
	}()
	go func() {
		_, err := io.Copy(b, a)
		errc <- err
	}()
	return <-errc
}
