/*
Represents a user in the system.

Users are managed outside of the system, but
their IDs needs to be synced between the source of truth
and this system to maintain referential integrity
across all applications.
*/
package users

type User struct {
	Id string
}

/*
Creates a new user with a given ID
*/
func NewUserWithId(id string) User {
	return User{
		Id: id,
	}
}

/*
Creates an empty User.

Used commonly when a function must return a User and
an error
*/
func Nil() User {
	return User{}
}
