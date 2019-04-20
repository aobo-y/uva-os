package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

type user struct {
	uname string
	psw   string
	port  string
}

var users []user

func main() {
	home := os.Getenv("HOME")
	dirpath := path.Join(home, ".hole_punch")
	fpath := path.Join(dirpath, "users")

	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(fpath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		panic(err)
	}

	sz := stat.Size()
	if sz != 0 {
		buf := make([]byte, sz)
		_, err = f.Read(buf)
		if err != nil {
			panic(err)
		}

		for _, line := range strings.Split(string(buf), "\n") {
			if line == "" {
				continue
			}
			tokens := strings.Split(line, " ")
			users = append(users, user{tokens[0], tokens[1], tokens[2]})
		}
	}

	var uname, pwd, port string
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Plese input username:")
	for {
		scanner.Scan()
		uname = scanner.Text()
		uname = strings.TrimSpace(uname)

		if uname == "" {
			fmt.Println("Username cannot be empty:")
			continue
		}

		if strings.Contains(uname, " ") {
			fmt.Println("Username cannot contain space:")
			continue
		}

		dup := false
		for i := 0; i < len(users); i++ {
			if users[i].uname == uname {
				dup = true
				break
			}
		}

		if dup {
			fmt.Println("Username exists, please input another one:")
			continue
		}

		break
	}

	fmt.Println("Plese input password:")
	for {
		scanner.Scan()
		pwd = scanner.Text()
		pwd = strings.TrimSpace(pwd)

		if pwd == "" {
			fmt.Println("Password cannot be empty:")
			continue
		}

		if strings.Contains(pwd, " ") {
			fmt.Println("Password cannot contain space:")
			continue
		}

		break
	}

	fmt.Println("Plese input port:")
	for {
		scanner.Scan()
		port = scanner.Text()
		port = strings.TrimSpace(port)

		if port == "" {
			fmt.Println("Port cannot be empty:")
			continue
		}

		_, err = strconv.Atoi(port)

		if err != nil {
			fmt.Println("Ivalid port number:")
			continue
		}

		dup := false
		for i := 0; i < len(users); i++ {
			if users[i].port == port {
				dup = true
				break
			}
		}

		if dup {
			fmt.Println("Port exists, please input another one:")
			continue
		}

		break
	}

	f.WriteString(uname + " " + pwd + " " + port + "\n")
	fmt.Println("Success!")
}
