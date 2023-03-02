package data

import (
	"database/sql"
	"fmt"
	"os"

	db2 "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"github.com/upper/db/v4/adapter/postgresql"
)

var db *sql.DB
var upper db2.Session

type Models struct {
	// any models inserted here (and in the New functions)
	// are easily accessible throughout the entire application
	Users  User
	Tokens Token
}

func New(databasePool *sql.DB) Models {
	db = databasePool

	if os.Getenv("DATABASE_TYPE") == "mysql" || os.Getenv("DATABASE_TYPE") == "mariadb" {
		// TODO: add mysql/mariadb models here
		upper, _ = mysql.New(db)
	} else if os.Getenv("DATABASE_TYPE") == "postgres" || os.Getenv("DATABASE_TYPE") == "postgresql" {
		upper, _ = postgresql.New(db)
	}

	return Models{
		Users:  User{},
		Tokens: Token{},
	}
}

func getInsertID(i db2.ID) int {

	if i == nil {
		return 0
	}

	idType := fmt.Sprintf("%T", i)

	if idType == "int64" {
		return int(i.(int64))
	} else if idType == "func() db.ID" {
		return getInsertID(i.(func() db2.ID)())
	}

	return int(i.(int))
}
