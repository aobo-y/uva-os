package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"punch"

	"golang.org/x/crypto/bcrypt"
)

func openConfig() (*os.File, error) {
	err := os.MkdirAll(punch.Dirpath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(punch.Fpath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0760)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func main() {
	users := *punch.ReadUsers()

	var uname, pswd, port string
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
			if users[i].Uname == uname {
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
		pswd = scanner.Text()
		pswd = strings.TrimSpace(pswd)

		if pswd == "" {
			fmt.Println("Password cannot be empty:")
			continue
		}

		if strings.Contains(pswd, " ") {
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
			if users[i].Port == port {
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

	hashedpswd, err := bcrypt.GenerateFromPassword([]byte(pswd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	pswd = string(hashedpswd)

	f, err := openConfig()
	if err != nil {
		panic(err)
	}

	f.WriteString(uname + " " + pswd + " " + port + "\n")
	fmt.Println("Success!")
}
