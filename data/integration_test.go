// go:build integration

// run tests with this command: go test . --tags integration -count=1
package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	host     = "localhost"
	user     = "postgres"
	password = "secret"
	dbName   = "gemquick_test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var dummyUser = User{
	FirstName: "John",
	LastName:  "Doe",
	Email:     "john@doe.com",
	Active:    1,
	Password:  "password",
}

var models Models
var testDB *sql.DB
var resource *dockertest.Resource
var pool *dockertest.Pool

func TestMain(m *testing.M) {
	os.Setenv("DATABASE_TYPE", "postgres")

	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	pool = p

	options := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_DB=" + dbName,
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {{HostIP: "0.0.0.0", HostPort: port}},
		},
	}

	resource, err = pool.RunWithOptions(&options)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("Could not start resource: %s", err)
	}

	if err = pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("Could not connect to docker: %s", err)
	}

	err = createTables(testDB)
	if err != nil {
		log.Fatal(err)
	}

	models = New(testDB)

	code := m.Run()

	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTables(db *sql.DB) error {
	stmt := `
	CREATE OR REPLACE FUNCTION trigger_set_timestamp()
	RETURNS TRIGGER AS $$
	BEGIN
	  NEW.updated_at = NOW();
	RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;
	
	drop table if exists users cascade;
	
	CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		first_name character varying(255) NOT NULL,
		last_name character varying(255) NOT NULL,
		user_active integer NOT NULL DEFAULT 0,
		email character varying(255) NOT NULL UNIQUE,
		password character varying(60) NOT NULL,
		created_at timestamp without time zone NOT NULL DEFAULT now(),
		updated_at timestamp without time zone NOT NULL DEFAULT now()
	);
	
	CREATE TRIGGER set_timestamp
		BEFORE UPDATE ON users
		FOR EACH ROW
		EXECUTE PROCEDURE trigger_set_timestamp();
	
	drop table if exists remember_tokens;
	
	CREATE TABLE remember_tokens (
		id SERIAL PRIMARY KEY,
		user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
		remember_token character varying(100) NOT NULL,
		created_at timestamp without time zone NOT NULL DEFAULT now(),
		updated_at timestamp without time zone NOT NULL DEFAULT now()
	);
	
	CREATE TRIGGER set_timestamp
		BEFORE UPDATE ON remember_tokens
		FOR EACH ROW
		EXECUTE PROCEDURE trigger_set_timestamp();
	
	drop table if exists tokens;
	
	CREATE TABLE tokens (
		id SERIAL PRIMARY KEY,
		user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
		first_name character varying(255) NOT NULL,
		email character varying(255) NOT NULL,
		token character varying(255) NOT NULL,
		token_hash bytea NOT NULL,
		created_at timestamp without time zone NOT NULL DEFAULT now(),
		updated_at timestamp without time zone NOT NULL DEFAULT now(),
		expiry timestamp without time zone NOT NULL
	);
	
	CREATE TRIGGER set_timestamp
		BEFORE UPDATE ON tokens
		FOR EACH ROW
		EXECUTE PROCEDURE trigger_set_timestamp();	
	`

	_, err := db.Exec(stmt)

	return err
}

func TestUser_Table(t *testing.T) {
	s := models.Users.Table()
	if s != "users" {
		t.Errorf("expected %s, got %s", "users", s)
	}
}

func TestUser_Insert(t *testing.T) {
	user, err := models.Users.Create(dummyUser)
	if err != nil {
		t.Error(err)
	}

	if user.ID == 0 {
		t.Errorf("expected id > 0, got %d", user.ID)
	}
}

func TestUser_Get(t *testing.T) {
	user, err := models.Users.Find(1)
	if err != nil {
		t.Error(err)
	}

	if user.ID == 0 {
		t.Errorf("expected id > 0, got %d", user.ID)
	}
}

func TestUser_All(t *testing.T) {
	users, err := models.Users.All()
	if err != nil {
		t.Error(err)
	}

	if len(users) == 0 {
		t.Errorf("expected len(users) > 0, got %d", len(users))
	}
}

func TestUser_ByEmail(t *testing.T) {
	user, err := models.Users.ByEmail("john@doe.com")
	if err != nil {
		t.Error(err)
	}

	if user.ID == 0 {
		t.Errorf("expected id > 0, got %d", user.ID)
	}
}

func TestUser_Update(t *testing.T) {
	user, err := models.Users.Find(1)
	if err != nil {
		t.Error(err)
	}

	user.FirstName = "Jane"
	user.LastName = "Doe"
	user.Email = "jane@doe.com"

	_, err = user.Update(*user)

	if err != nil {
		t.Error(err)
	}

	user, err = models.Users.Find(1)
	if err != nil {
		t.Error(err)
	}

	if user.FirstName != "Jane" {
		t.Errorf("expected %s, got %s", "Jane", user.FirstName)
	}

	if user.LastName != "Doe" {
		t.Errorf("expected %s, got %s", "Doe", user.LastName)
	}

	if user.Email != "jane@doe.com" {
		t.Errorf("expected %s, got %s", "jane@doe.com", user.Email)
	}
}
