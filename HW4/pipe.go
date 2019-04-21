// Utils to exchange data between connections
package punch

import (
	"net"
)

// Keep reading from one connection and writing to another one
// Allow giving an optional channel to send the number of bytes streamed inn realtime
// Stop when either conn gives error and return the error
func Stream(from net.Conn, to net.Conn, chn chan int) error {
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

// Stream data between two connections interactively
// Allowing giving two optional channels in the arguments
// to receive the realtime number of bytes streamed of each direction respectively
// Stops when either conn gives error and returnthe error
func Pipe(conn1 net.Conn, conn2 net.Conn, chn1 chan int, chn2 chan int) error {
	done := make(chan error)

	go func() {
		done <- Stream(conn1, conn2, chn1)
	}()

	go func() {
		done <- Stream(conn2, conn1, chn2)
	}()

	return <-done
}
