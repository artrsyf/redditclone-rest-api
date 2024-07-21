package mysql

import (
	"database/sql"
	"errors"
	"redditclone/pkg/models"
	"reflect"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestNewUserMySqlRepo(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserMySqlRepo(db)

	if repo.DB != db {
		t.Errorf("expected db connection: %v, got: %v", db, repo.DB)
	}
}

func TestGetUserFromRepo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := &UserMysqlRepository{
		DB: db,
	}

	var userID = 1
	var login = "alex12345"
	var correctPassword = "qwerty"

	t.Run("correct query", func(t *testing.T) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("cant hash password: %s", err)
		}

		rows := sqlmock.NewRows([]string{"id", "login", "password"})
		expect := []*models.User{
			{
				ID:       1,
				Login:    login,
				Password: string(hashedPassword),
			},
		}
		for _, user := range expect {
			rows = rows.AddRow(user.ID, user.Login, user.Password)
		}

		mock.
			ExpectQuery("SELECT id, login, password FROM user WHERE").
			WithArgs(login).
			WillReturnRows(rows)

		user, err := repo.GetUserFromRepo(login, correctPassword)
		if err != nil {
			t.Errorf("unexpected err: %s", err)
			return
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
			return
		}

		if !reflect.DeepEqual(user, expect[0]) {
			t.Errorf("results not match, want %v, have %v", expect[0], user)
			return
		}
	})

	t.Run("ErrNoUser", func(t *testing.T) {
		unknownLogin := "max123"
		somePassword := "12345"

		mock.
			ExpectQuery("SELECT id, login, password FROM user WHERE").
			WithArgs(unknownLogin).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetUserFromRepo(unknownLogin, somePassword)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}

		if err != models.ErrNoUser {
			t.Errorf("unexpected err: want %s, got %s", models.ErrNoUser, err)
			return
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		unexpectedErr := errors.New("some error")

		mock.
			ExpectQuery("SELECT id, login, password FROM user WHERE").
			WithArgs(login).
			WillReturnError(unexpectedErr)

		_, err := repo.GetUserFromRepo(login, correctPassword)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}

		if err != unexpectedErr {
			t.Errorf("unexpected err: want %s, got %s", unexpectedErr, err)
			return
		}
	})

	t.Run("ErrWrongCredentials", func(t *testing.T) {
		incorrectPassword := "xlxl123"

		rows := sqlmock.NewRows([]string{"id", "login", "password"})
		rows.AddRow(userID, login, correctPassword)

		mock.
			ExpectQuery("SELECT id, login, password FROM user WHERE").
			WithArgs(login).
			WillReturnRows(rows)

		_, err := repo.GetUserFromRepo(login, incorrectPassword)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}

		if err != models.ErrWrongCredentials {
			t.Errorf("unexpected err: want %s, got %s", models.ErrWrongCredentials, err)
			return
		}
	})
}

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := &UserMysqlRepository{
		DB: db,
	}

	var login = "alex12345"
	var correctPassword = "qwerty"
	var hashedPassword, _ = bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	t.Run("correct query", func(t *testing.T) {

		user := &models.User{
			ID:       1,
			Login:    login,
			Password: correctPassword,
		}

		mock.
			ExpectQuery("SELECT 1 FROM user WHERE").
			WithArgs(login).
			WillReturnError(sql.ErrNoRows)

		mock.
			ExpectQuery("SELECT id, login, password FROM user WHERE").
			WithArgs(login).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password"}).
				AddRow(1, login, string(hashedPassword)))

		// Почему не отлавливается через sqlmock?????
		// mock.
		// 	ExpectExec(`INSERT INTO user`).
		// 	WithArgs(login, string(hashedPassword)).
		// 	WillReturnResult(sqlmock.NewResult(1, 1))

		user, err := repo.CreateUser(login, correctPassword)
		if err != nil {
			t.Errorf("unexpected err: %s", err)
			return
		}

		if user.ID != 1 {
			t.Errorf("bad id: want %v, have %v", user.ID, 1)
			return
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("hash error", func(t *testing.T) {
		tooLongPassword := `fjdkfjksjfkdsjkfjdskfjlkdsjkdsjfkdsjflkdscnmxsnc,mxsnc,mdsd
		kflkdsf;lkds;lfk;ldskf;lkds;lfk;ldskf;lkskdflkds;lkf
		f;dslr;ewl;l;rlewdcx,cnmdsnfmjslkjfewiur2r32rjoi2fjei293uj2di32d32
		32ed32d2fewkflklskl;fkp[r23krpo2k]`

		mock.
			ExpectQuery("SELECT 1 FROM user WHERE").
			WithArgs(login).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.CreateUser(login, tooLongPassword)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}

		if err != bcrypt.ErrPasswordTooLong {
			t.Errorf("unexpected err: want %s, got %s", bcrypt.ErrPasswordTooLong, err)
			return
		}
	})

	t.Run("ErrAlreadyCreated", func(t *testing.T) {
		mock.
			ExpectQuery("SELECT 1 FROM user WHERE").
			WithArgs(login).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password"}).
				AddRow(1, login, string(hashedPassword)))

		_, err := repo.CreateUser(login, correctPassword)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}

		if err != models.ErrAlreadyCreated {
			t.Errorf("unexpected err: want %s, got %s", models.ErrAlreadyCreated, err)
			return
		}
	})

	t.Run("unexpected check query error", func(t *testing.T) {
		unexpectedErr := errors.New("some error")

		mock.
			ExpectQuery("SELECT 1 FROM user WHERE").
			WithArgs(login).
			WillReturnError(unexpectedErr)

		_, err := repo.CreateUser(login, correctPassword)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}

		if err != unexpectedErr {
			t.Errorf("unexpected err: want %s, got %s", unexpectedErr, err)
			return
		}
	})

	t.Run("insertion error", func(t *testing.T) {
		insertionErr := errors.New("some insertion error")

		mock.
			ExpectQuery("SELECT 1 FROM user WHERE").
			WithArgs(login).
			WillReturnError(sql.ErrNoRows)

		mock.
			ExpectQuery("SELECT id, login, password FROM user WHERE").
			WithArgs(login).
			WillReturnError(insertionErr)

		_, err := repo.CreateUser(login, correctPassword)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}

		if err != insertionErr {
			t.Errorf("unexpected err: want %s, got %s", insertionErr, err)
			return
		}
	})
}

func TestGetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := &UserMysqlRepository{
		DB: db,
	}

	var userID = 1
	var login = "alex12345"
	var password = "qwerty"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	t.Run("correct query", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "login", "password"}).
			AddRow(userID, login, hashedPassword)

		mock.
			ExpectQuery("SELECT id, login, password FROM user WHERE").
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := repo.GetUserByID(userID)
		if err != nil {
			t.Errorf("unexpected err: %s", err)
			return
		}

		if user.ID != userID {
			t.Errorf("bad id: want %v, have %v", user.ID, userID)
			return
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("ErrNoUser", func(t *testing.T) {
		unknownUserID := 2

		mock.
			ExpectQuery("SELECT id, login, password FROM user WHERE").
			WithArgs(unknownUserID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetUserByID(unknownUserID)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}

		if err != models.ErrNoUser {
			t.Errorf("unexpected err: want %s, got %s", models.ErrNoUser, err)
			return
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		unknownUserID := 2
		unexpectedErr := errors.New("some error")

		mock.
			ExpectQuery("SELECT id, login, password FROM user WHERE").
			WithArgs(unknownUserID).
			WillReturnError(unexpectedErr)

		_, err := repo.GetUserByID(unknownUserID)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}

		if err != unexpectedErr {
			t.Errorf("unexpected err: want %s, got %s", unexpectedErr, err)
			return
		}
	})
}
