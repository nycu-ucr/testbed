package context

import (
	"sync"
)

type User struct {
	Id       string `json:"UserId"`
	Name     string `json:"UserName"`
	Password string `json:"UserPassword"`
}

type Users []User

var UserPool = new(sync.Map)

func UserFindById(id string) (*User, bool) {
	if value, ok := UserPool.Load(id); ok {
		return value.(*User), ok
	}

	return nil, false
}

func AddUserToUserPool(user *User) {
	id := user.Id
	UserPool.Store(id, user)
}

func DeleteUserFromUserPool(id string) {
	UserPool.Delete(id)
}
