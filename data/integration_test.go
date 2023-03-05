// go:build integration

// run tests with this command: go test . --tags integration -count=1
package data

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

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
	os.Setenv("UPPER_DB_LOG", "ERROR")

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

func TestUser_PasswordMatches(t *testing.T) {
	user, err := models.Users.Find(1)
	if err != nil {
		t.Error(err)
	}

	matches, err := user.PasswordMatches("password")
	if err != nil {
		t.Error("error checking password match", err)
	}

	if !matches {
		t.Error("expected passwords to match, but got false")
	}

	matches, err = user.PasswordMatches("wrongpassword")
	if err != nil {
		t.Error("error checking password match", err)
	}

	if matches {
		t.Error("expected passwords to not match, but got true")
	}
}

func TestUser_ResetPassword(t *testing.T) {
	err := models.Users.ResetPassword(1, "newpassword")
	if err != nil {
		t.Error("error resetting password", err)
	}

	err = models.Users.ResetPassword(2, "newpassword")
	if err == nil {
		t.Error("did not get an error when resetting a password for a non existent user", err)
	}
}

func TestUser_Delete(t *testing.T) {
	user, err := models.Users.Find(1)
	if err != nil {
		t.Error(err)
	}

	err = user.Delete(user.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestToken_Table(t *testing.T) {
	s := models.Tokens.Table()
	if s != "tokens" {
		t.Errorf("expected %s, got %s", "tokens", s)
	}
}

func TestToken_GenerateToken(t *testing.T) {
	user, err := models.Users.Create(dummyUser)
	if err != nil {
		t.Error(err)
	}

	_, err = models.Tokens.GenerateToken(user.ID, time.Hour*24*365)
	if err != nil {
		t.Error("error generating token", err)
	}
}

func TestToken_Insert(t *testing.T) {
	user, err := models.Users.ByEmail(dummyUser.Email)
	if err != nil {
		t.Error(err)
	}

	token, err := models.Tokens.GenerateToken(user.ID, time.Hour*24*365)
	if err != nil {
		t.Error("error generating token", err)
	}

	err = models.Tokens.Insert(*token, *user)
	if err != nil {
		t.Error("error creating token", err)
	}
}

func TestToken_GetUserForToken(t *testing.T) {
	token := "abc"
	_, err := models.Tokens.GetUserForToken(token)
	if err == nil {
		t.Error("expected an error, got nil")
	}

	user, err := models.Users.ByEmail(dummyUser.Email)
	if err != nil {
		t.Error(err)
	}

	_, err = models.Tokens.GetUserForToken(user.Token.PlainText)
	if err != nil {
		t.Error("error getting user for token", err)
	}

	_, err = models.Tokens.GetUserForToken("wrongtoken")
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestToken_GetTokensForUser(t *testing.T) {
	user, err := models.Users.ByEmail(dummyUser.Email)
	if err != nil {
		t.Error(err)
	}

	tokens, err := models.Tokens.GetTokensForUser(user.ID)
	if err != nil {
		t.Error("error getting tokens for user", err)
	}

	if len(tokens) == 0 {
		t.Error("expected tokens, got none")
	}
}

func TestToken_GetByToken(t *testing.T) {
	user, err := models.Users.ByEmail(dummyUser.Email)
	if err != nil {
		t.Error(err)
	}

	_, err = models.Tokens.GetByToken(user.Token.PlainText)
	if err != nil {
		t.Error("error getting token by token", err)
	}

	_, err = models.Tokens.GetByToken("wrongtoken")
	if err == nil {
		t.Error("expected an error when looking for a non existent token, but got nil")
	}
}

func TestToken_Get(t *testing.T) {
	user, err := models.Users.ByEmail(dummyUser.Email)
	if err != nil {
		t.Error(err)
	}

	_, err = models.Tokens.Find(user.Token.ID)
	if err != nil {
		t.Error("error finding token", err)
	}

}

var authData = []struct {
	name          string
	token         string
	email         string
	errorExpected bool
	message       string
}{
	{"invalid", "abcdefghijklmnopqrstuvwxyz", "not@existing.com", true, "invalid token accepted as valid"},
	{"invalid_length", "abcdefghijklmnopqrstuvwxy", "non@existing.com", true, "invalid token length accepted as valid"},
	{"no_user", "abcdefghijklmnopqrstuvwxyz", "non@existing.com", true, "token accepted for non existing user"},
	{"valid", "abcdefghijklmnopqrstuvwxyz", dummyUser.Email, false, "valid token reported as invalid"},
}

func TestToken_AuthenticateToken(t *testing.T) {
	for _, data := range authData {
		token := ""
		if data.email == dummyUser.Email {
			user, err := models.Users.ByEmail(data.email)
			if err != nil {
				t.Error("failed to get user: ", err)
			}
			token = user.Token.PlainText
		} else {
			token = data.token
		}

		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		_, err := models.Tokens.AuthenticateToken(req)
		if err == nil && data.errorExpected {
			t.Errorf("%s: %s", data.name, data.message)
		} else if err != nil && !data.errorExpected {
			t.Errorf("%s: %s - %s", data.name, data.message, err)
		} else {
			t.Logf("passed: %s", data.name)
		}
	}
}

func TestToken_Delete(t *testing.T) {
	user, err := models.Users.ByEmail(dummyUser.Email)
	if err != nil {
		t.Error(err)
	}

	// Get first user token
	tokens, err := models.Tokens.GetTokensForUser(user.ID)
	if err != nil {
		t.Error("error getting tokens for user", err)
	}

	err = models.Tokens.Delete(tokens[0].ID)
	if err != nil {
		t.Error("error deleting token", err)
	}
}

func TestToken_ExpiredToken(t *testing.T) {
	user, err := models.Users.ByEmail(dummyUser.Email)
	if err != nil {
		t.Error(err)
	}

	// Get first user token
	tokens, err := models.Tokens.GenerateToken(user.ID, -time.Hour)
	if err != nil {
		t.Error("error getting tokens for user", err)
	}

	err = models.Tokens.Insert(*tokens, *user)
	if err != nil {
		t.Error("error creating token", err)
	}

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.PlainText)

	_, err = models.Tokens.AuthenticateToken(req)
	if err == nil {
		t.Error("expired token accepted as valid")
	}
}

func TestToken_BadHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)

	_, err := models.Tokens.AuthenticateToken(req)
	if err == nil {
		t.Error("missing header accepted as valid")
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "abc")

	_, err = models.Tokens.AuthenticateToken(req)
	if err == nil {
		t.Error("invalid header accepted as valid")
	}

	newUser := User{
		Email:     "jimmie@developer.com",
		Password:  "password",
		FirstName: "Jimmie",
		LastName:  "Developer",
		Active:    1,
	}

	user, err := models.Users.Create(newUser)
	if err != nil {
		t.Error(err)
	}

	token, err := models.Tokens.GenerateToken(user.ID, 1*time.Hour)
	if err != nil {
		t.Error("error generating token", err)
	}

	err = models.Tokens.Insert(*token, *user)
	if err != nil {
		t.Error("error creating token", err)
	}

	err = models.Users.Delete(user.ID)
	if err != nil {
		t.Error("error deleting user", err)
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token.PlainText)

	_, err = models.Tokens.AuthenticateToken(req)
	if err == nil {
		t.Error("deleted user token accepted as valid")
	}
}

func TestToken_DeleteNonExistentToken(t *testing.T) {
	err := models.Tokens.DeleteByToken("avc")
	if err != nil {
		t.Error("error deleting token")
	}

	err = models.Tokens.Delete(999999)
	if err != nil {
		t.Error("error deleting token")
	}
}

func TestToken_ValidToken(t *testing.T) {
	user, err := models.Users.ByEmail(dummyUser.Email)
	if err != nil {
		t.Error(err)
	}

	// Generate token for user
	token, err := models.Tokens.GenerateToken(user.ID, 1*time.Hour)
	if err != nil {
		t.Error("error generating token", err)
	}

	err = models.Tokens.Insert(*token, *user)
	if err != nil {
		t.Error("error creating token", err)
	}

	okay, err := models.Tokens.ValidateToken(token.PlainText)
	if err != nil {
		t.Error("error validating token", err)
	}

	if !okay {
		t.Error("valid token reported as not valid")
	}

	okay, _ = models.Tokens.ValidateToken("abc")
	if okay {
		t.Error("invalid token reported as valid")
	}

	user, err = models.Users.ByEmail(dummyUser.Email)
	if err != nil {
		t.Error(err)
	}

	err = models.Tokens.Delete(user.Token.ID)
	if err != nil {
		t.Error("error deleting token", err)
	}

	okay, _ = models.Tokens.ValidateToken(user.Token.PlainText)

	if okay {
		t.Error("deleted token reported as valid")
	}
}
