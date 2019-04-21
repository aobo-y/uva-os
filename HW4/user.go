// Utils to get user config
package punch

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type User struct {
	Uname string
	Pswd  string
	Port  string
}

var home = os.Getenv("HOME")
var Dirpath = path.Join(home, ".hole_punch")
var Fpath = path.Join(Dirpath, "users")

// Read the config and return an array of users
func ReadUsers() *[]User {
	var users []User

	content, err := ioutil.ReadFile(Fpath)
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
