package mysql

import (
	"database/sql"
	"redditclone/pkg/models"

	"golang.org/x/crypto/bcrypt"
)

type UserMysqlRepository struct {
	DB *sql.DB
}

func NewUserMySqlRepo(db *sql.DB) *UserMysqlRepository {
	return &UserMysqlRepository{
		DB: db,
	}
}

func (repo *UserMysqlRepository) GetUserFromRepo(login, pass string) (*models.User, error) {
	user := &models.User{}

	err := repo.DB.
		QueryRow("SELECT id, login, password FROM user WHERE login = ?", login).
		Scan(&user.ID, &user.Login, &user.Password)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoUser
	} else if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass)); err != nil {
		return nil, models.ErrWrongCredentials
	}

	return user, nil
}

func (repo *UserMysqlRepository) CreateUser(login, pass string) (*models.User, error) {
	err := repo.DB.
		QueryRow("SELECT 1 FROM user WHERE login = ?", login).Scan(new(int))
	if err == sql.ErrNoRows {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		repo.DB.Exec(
			"INSERT INTO user (`login`, `password`) VALUES (?, ?)",
			login,
			string(hashedPassword),
		)

		user := &models.User{}
		err = repo.DB.
			QueryRow("SELECT id, login, password FROM user WHERE login = ?", login).
			Scan(&user.ID, &user.Login, &user.Password)
		if err != nil {
			return nil, err
		}

		return user, nil
	}

	if err != nil {
		return nil, err
	}

	return nil, models.ErrAlreadyCreated
}

func (repo *UserMysqlRepository) GetUserByID(userID int) (*models.User, error) {
	user := &models.User{}

	err := repo.DB.
		QueryRow("SELECT id, login, password FROM user WHERE id = ?", userID).
		Scan(&user.ID, &user.Login, &user.Password)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoUser
	} else if err != nil {
		return nil, err
	}

	return user, nil
}
