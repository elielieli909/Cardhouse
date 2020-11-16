package users

// Users will hold all the users, through two maps: Users ([id]*User), and IDs ([name]id)
type Users struct {
	// Users is a map of User's keyed off User.id
	users map[int]*User

	// IDs is a map of User.id's keyed off User.name
	IDs map[string]int
}

// NewUsers is a constructor for the Users struct
func NewUsers() *Users {
	us := new(Users)
	us.users = make(map[int]*User)
	us.IDs = make(map[string]int)
	return us
}

// NewUser adds a new User to the struct and returns the new User
func (us *Users) NewUser() *User {
	u := createUser()
	us.users[u.id] = u
	us.IDs[u.name] = u.id

	return u
}
