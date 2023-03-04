package data

import (
	"fmt"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	db2 "github.com/upper/db/v4"
)

func TestNew(t *testing.T) {
	fakeDB, _, _ := sqlmock.New()
	defer fakeDB.Close()

	_ = os.Setenv("DATABASE_TYPE", "mysql")
	m := New(fakeDB)

	if fmt.Sprintf("%T", m) != "data.Models" {
		t.Error("Wrong type", fmt.Sprintf("%T", m))
	}
}

func TestGetInsertID(t *testing.T) {
	var id db2.ID
	id = int64(1)

	returnedId := getInsertID(id)

	if fmt.Sprintf("%T", returnedId) != "int" {
		t.Error("Wrong type", fmt.Sprintf("%T", returnedId))
	}

	id = 1
	returnedId = getInsertID(id)

	if fmt.Sprintf("%T", returnedId) != "int" {
		t.Error("Wrong type", fmt.Sprintf("%T", returnedId))
	}

}
