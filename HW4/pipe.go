package punch

import (
	"net"
)

func stream(from net.Conn, to net.Conn, chn chan int) error {
	buf := make([]byte, 1024)

	for {
		n, err := from.Read(buf)

		if err != nil {
			return err
		}
		_, err = to.Write(buf[:n])

		if err != nil {
			return err
		}

		if chn != nil {
			chn <- n
		}
	}
}

func Pipe(conn1 net.Conn, conn2 net.Conn, chn1 chan int, chn2 chan int) error {
	done := make(chan error)

	go func() {
		done <- stream(conn1, conn2, chn1)
	}()

	go func() {
		done <- stream(conn2, conn1, chn2)
	}()

	return <-done
}
