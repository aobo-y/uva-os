package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	uname string
	psw   string
	port  string
}

var home = os.Getenv("HOME")
var dirpath = path.Join(home, ".hole_punch")
var fpath = path.Join(dirpath, "users")

func ReadUser() *[]User {
	var users []User

	content, err := ioutil.ReadFile(fpath)
	if err == nil && len(content) > 0 {
		for _, line := range strings.Split(string(content), "\n") {
			if line == "" {
				continue
			}
			tokens := strings.Split(line, " ")
			users = append(users, User{tokens[0], tokens[1], tokens[2]})
		}
	}

	return &users
}

func openConfig() (*os.File, error) {
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(fpath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0760)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func main() {
	users := *ReadUser()

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

		_, err := strconv.Atoi(port)

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

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	pwd = string(hashedPwd)

	f, err := openConfig()
	if err != nil {
		panic(err)
	}

	f.WriteString(uname + " " + pwd + " " + port + "\n")
	fmt.Println("Success!")
}
