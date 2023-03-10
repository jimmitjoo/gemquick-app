package data

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jimmitjoo/gemquick"
	up "github.com/upper/db/v4"
)

type User struct {
	ID        int       `db:"id,omitempty"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Active    int       `db:"user_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Token     Token     `db:"-"`
}

func (u *User) Table() string {
	return "users"
}

func (u *User) Validate(validator *gemquick.Validation) {
	validator.Check(u.FirstName != "", "first_name", "this field is required")
	validator.Check(u.LastName != "", "last_name", "this field is required")
	validator.IsEmail("email", u.Email)
}

func (u *User) All() ([]*User, error) {
	collection := upper.Collection(u.Table())

	var users []*User

	res := collection.Find().OrderBy("created_at desc")
	err := res.All(&users)

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (u *User) Find(id int) (*User, error) {
	collection := upper.Collection(u.Table())

	var user User

	res := collection.Find(up.Cond{"id =": id})
	err := res.One(&user)

	if err != nil {
		return nil, err
	}

	var token Token

	collection = upper.Collection(token.Table())
	res = collection.Find(up.Cond{"user_id =": user.ID, "expiry >": time.Now()}).OrderBy("created_at desc").Limit(1)
	err = res.One(&token)

	if err != nil {
		if err != up.ErrNoMoreRows && err != up.ErrNilRecord {
			return nil, err
		}
	}

	user.Token = token

	return &user, nil
}

func (u *User) ByEmail(email string) (*User, error) {
	collection := upper.Collection(u.Table())

	var user User

	res := collection.Find(up.Cond{"email =": email})
	err := res.One(&user)

	if err != nil {
		return nil, err
	}

	var token Token
	collection = upper.Collection(token.Table())
	res = collection.Find(up.Cond{"user_id =": user.ID, "expiry >": time.Now()}).OrderBy("created_at desc").Limit(1)
	err = res.One(&token)

	if err != nil {
		if err != up.ErrNoMoreRows && err != up.ErrNilRecord {
			return nil, err
		}
	}

	user.Token = token

	return &user, nil
}

func (u *User) Update(user User) (*User, error) {
	collection := upper.Collection(u.Table())

	user.UpdatedAt = time.Now()

	res := collection.Find(up.Cond{"id =": user.ID})
	err := res.Update(user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *User) Create(user User) (*User, error) {
	collection := upper.Collection(u.Table())

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)

	if err != nil {
		return nil, err
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Password = string(hashedPassword)

	res, err := collection.Insert(user)

	if err != nil {
		return nil, err
	}

	user.ID = getInsertID(res.ID)

	return &user, nil
}

func (u *User) Delete(id int) error {
	collection := upper.Collection(u.Table())

	res := collection.Find(up.Cond{"id =": id})
	err := res.Delete()

	if err != nil {
		return err
	}

	return nil
}

func (u *User) ResetPassword(id int, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)

	if err != nil {
		return err
	}

	user, err := u.Find(id)

	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)

	_, err = u.Update(*user)

	if err != nil {
		return err
	}

	return nil
}

func (u *User) PasswordMatches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainText))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
