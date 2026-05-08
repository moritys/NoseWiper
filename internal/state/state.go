package state

type UserState struct {
	State string
}

var Users = make(map[int64]*UserState)
