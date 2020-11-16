package users

var curUID int = 0

// User is the base type representing a user; has an id, cash balance, name, array of owned assets, and a map of assetID to shares owned.
type User struct {
	id int
	// Cash on hand
	cash int
	name string
	// TODO: add metadata

	// Array of Asset.id's representing this user's owned assets
	assets []int
	// sharesOwned[assetID] = number of shares owned
	sharesOwned map[int]int
}

// createUser returns a new user object; to be used by users.go internally
func createUser() *User {
	curUID++
	u := new(User)
	u.id = curUID
	u.cash = 0
	u.assets = make([]int, 0)
	u.sharesOwned = make(map[int]int)
	return u
}

// DepositCash adds amount to u's balance
func (u *User) DepositCash(amount int) int {
	if amount <= 0 {
		return u.cash
	}

	u.cash += amount
	return u.cash
}

// WithdrawCash removes amount from u's balance
func (u *User) WithdrawCash(amount int) int {
	if amount <= 0 {
		return u.cash
	}

	if amount > u.cash {
		return u.cash
	}

	u.cash -= amount
	return u.cash
}
