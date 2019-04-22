package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"sync"
	"time"

	"punch"

	"golang.org/x/crypto/bcrypt"
)

var users []punch.User

type connection struct {
	sync.RWMutex        // mutex for concurrent read & write
	uname        string // username
	owport       string // outward facing port
	cliport      string // punch client facing port
	cliip        string // client ip address
	bytrcv       int    // bytes received
	bytsnd       int    // bytes sent
}

var openConns = make(map[string]*connection) // maps of open connections

var optSplit = regexp.MustCompile(`\s+`) // regex to split punch client input

// Pipe streams between the outward connection and punch client connection
// On-the-fly update the bytes received & sent
// Close both connections when done
func pipe(owconn net.Conn, cliconn net.Conn) {
	pipeLog := fmt.Sprintf("%v -> %v <--> %v <- %v", owconn.RemoteAddr().String(), owconn.LocalAddr().String(), cliconn.LocalAddr().String(), cliconn.RemoteAddr().String())

	fmt.Println("Pipe " + pipeLog)

	_, port, _ := net.SplitHostPort(owconn.LocalAddr().String())
	connEntry := openConns[port]

	done := make(chan error)
	rcvChn := make(chan int)
	sndChn := make(chan int)

	go func() {
		done <- punch.Pipe(owconn, cliconn, rcvChn, sndChn)
	}()

	for {
		end := false
		select {
		case n := <-rcvChn:
			connEntry.Lock()
			connEntry.bytrcv += n
			connEntry.Unlock()

		case n := <-sndChn:
			connEntry.Lock()
			connEntry.bytsnd += n
			connEntry.Unlock()

		case <-done:
			end = true
		}

		if end {
			break
		}
	}

	owconn.Close()
	cliconn.Close()

	fmt.Println("Close " + pipeLog)
}

// Open the specified outward facing port & Open a client facing port
// For each incoming conn, send nounce to punch client through control conn
// Wait for a client conn with the nounce for at most 10s
// If timeout, close the opened ports and end
// Else, pipe the outward conn & client conn
func listen(ctrconn net.Conn, port string, uname string) {
	l, err := net.Listen("tcp", ":"+port)

	if err != nil {
		log.Fatal("open outward facing port error: ", err)
		ctrconn.Write([]byte("FAIL"))
		ctrconn.Close()
		return
	}

	defer l.Close()
	fmt.Println("Listening to outward facing port:", port)

	clilstner, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("open punch client port error: ", err)
		return
	}

	defer clilstner.Close()

	_, cliport, _ := net.SplitHostPort(clilstner.Addr().String())
	fmt.Println("Listening to client facing port:", uname, cliport)

	cliip, _, _ := net.SplitHostPort(ctrconn.RemoteAddr().String())
	openConns[port] = &connection{
		uname:   uname,
		owport:  port,
		cliport: cliport,
		cliip:   cliip,
	}
	defer delete(openConns, port)

	cliconnChn := make(chan net.Conn) // client conn channel
	done := make(chan error)

	var nounce string

	// gorountine to wait for connection with correct nounce
	go func() {
		for {
			cliconn, err := clilstner.Accept()
			if err != nil {
				done <- fmt.Errorf("accept punch client connection error: %v", err)
				return
			}

			buf := make([]byte, 512)
			n, err := cliconn.Read(buf)
			if err != nil {
				done <- fmt.Errorf("read punch client connection error: %v", err)
				return
			}

			clinounce := string(buf[:n])

			if clinounce == nounce {
				cliconnChn <- cliconn
			}
		}
	}()

	// gorountine to monitor if the control connection is available
	// end the client handling if lose the control conn
	go func() {
		for {
			buf := make([]byte, 256)
			if _, err := ctrconn.Read(buf); err != nil {
				done <- fmt.Errorf("Control connection closed")
				return
			}
		}
	}()

	// gorountine to wait for outward connections & send invites
	// then wait for client connection within 10s
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				done <- err
				return
			}

			fmt.Println("Outward " + conn.RemoteAddr().String() + " -> " + conn.LocalAddr().String())

			nounce = randString(256)
			_, err = ctrconn.Write([]byte("CONNECT " + cliport + " " + nounce))
			if err != nil {
				// control conn has closed in the client side
				done <- fmt.Errorf("Write control connection error")
				return
			}

			// timeout after 10s
			select {
			case cliconn, ok := <-cliconnChn:
				if !ok {
					return
				}
				fmt.Println("Client " + cliconn.RemoteAddr().String() + " -> " + cliconn.LocalAddr().String())
				go pipe(conn, cliconn)
			case <-time.After(10 * time.Second):
				done <- fmt.Errorf("Punch client connection timeout")
				return
			}
		}
	}()

	err = <-done
	close(cliconnChn)
	fmt.Println("End open port for "+uname+":", err)
}

// Write the formatted realtime info of the openned connections
// to punch client through the control conn
func list(ctrconn net.Conn) {
	var res string

	for _, cEty := range openConns {
		line := fmt.Sprintf("%v %v %v %v", cEty.uname, cEty.owport, cEty.cliport, cEty.cliip)

		for _, byt := range []int{cEty.bytrcv, cEty.bytsnd} {
			val := float32(byt)
			var unit string

			for _, unit = range []string{"B", "KB", "MB", "GB"} {
				if val < 1024 {
					break
				}
				val /= 1024
			}

			line += fmt.Sprintf(" %.1f%v", val, unit)
		}

		res += line + "\n"
	}

	ctrconn.Write([]byte(res))
}

// Verify the username, password and port in OPEN cmd
// The username, password, and port should match the existing config
// The port should have not been openned by other punch client
func verify(uname string, pw string, port string) error {
	pwbyt := []byte(pw)

	for _, user := range users {
		if user.Uname != uname {
			continue
		}

		if user.Port != port {
			return fmt.Errorf("Invalid port")
		}

		hpswd := []byte(user.Pswd)
		err := bcrypt.CompareHashAndPassword(hpswd, pwbyt)

		if err != nil {
			return fmt.Errorf("Invalid password")
		}

		if _, in := openConns[port]; in {
			return fmt.Errorf("Port has openned")
		}

		return nil
	}

	return fmt.Errorf("Invalid username")
}

// Handle the incoming punch client connection
// Wait for only one the client request, either OPEN or  LIST
// Close the client conn in the end
func handleClient(conn net.Conn) {
	clistr := conn.RemoteAddr().String()
	fmt.Println("Received connection from " + clistr)

	buf := make([]byte, 512)
	nb, err := conn.Read(buf)

	if err != nil {
		fmt.Println("Lost connection from " + clistr)
		return
	}

	cmd := string(buf[:nb])
	tokens := optSplit.Split(cmd, -1)

	switch tokens[0] {
	case "OPEN":
		if len(tokens) != 4 {
			fmt.Println("Invalid OPEN operation from " + clistr)
			return
		}

		fmt.Println("Received OPEN command from "+clistr+":", tokens[1], tokens[3])

		err = verify(tokens[1], tokens[2], tokens[3])
		if err != nil {
			fmt.Println("Auth punch client "+clistr+" error:", err)

			conn.Write([]byte("FAIL"))
			conn.Close()
			return
		}

		listen(conn, tokens[3], tokens[1])

	case "LIST":
		if len(tokens) != 1 {
			fmt.Println("Invalid LIST operation from " + clistr)
			return
		}

		fmt.Println("Received LIST command from " + clistr)
		list(conn)

	default:
		fmt.Println("Invalid operation from " + clistr)
	}

	conn.Close()
}

// Start the punch server
// Support an optional argument for the port to listen to
func main() {
	port := "9999"

	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	l, err := net.Listen("tcp", ":"+port)

	if err != nil {
		log.Fatal("listen error: ", err)
	}

	fmt.Println("Punch server is listening at port " + port)

	users = *punch.ReadUsers()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error: ", err)
		}

		go handleClient(conn)
	}
}
