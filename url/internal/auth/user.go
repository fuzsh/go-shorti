package auth

type User struct {
	ID         int    `db:"user_id" json:"-"`
	Username   string `db:"username" json:"username"`
	Password   string `db:"password" json:"password"`
	IsVerified bool   `db:"is_verified" json:"isVerified"`
}
