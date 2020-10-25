package user

import (
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"strings"
	"time"

	"github.com/Erexo/Ventana/core/entity"
	"github.com/Erexo/Ventana/infrastructure/config"
	"github.com/Erexo/Ventana/infrastructure/db"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/pbkdf2"
)

const (
	minUsernameLength = 4
	minPasswordLength = 6

	saltSize   = 24
	hashSize   = 24
	iterations = 1000
)

var (
	ErrInvalidCredentials = errors.New("Invalid credentials")
)

type Service struct {
}

func CreateService() *Service {
	return &Service{}
}

type LoginInfo struct {
	AccessToken string      `json:"accesstoken"`
	Role        entity.Role `json:"role"`
}

func (s *Service) Login(username, password string) (LoginInfo, error) {
	conn, err := db.GetConnection()
	if err != nil {
		return LoginInfo{}, err
	}
	defer conn.Close()

	var dbpassword string
	var dbsalt *string
	var role entity.Role
	err = conn.QueryRow("SELECT password, salt, role FROM user WHERE username LIKE ?", username).Scan(&dbpassword, &dbsalt, &role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LoginInfo{}, ErrInvalidCredentials
		}
		return LoginInfo{}, err
	}

	if dbsalt != nil {
		var err error
		password, err = getHash(password, *dbsalt)
		if err != nil {
			return LoginInfo{}, err
		}
	}
	if password != dbpassword {
		return LoginInfo{}, ErrInvalidCredentials
	}

	now := time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":  math.MaxInt64,
		"nbf":  now,
		"iat":  now,
		"sub":  username,
		"role": role,
	})
	config := config.GetConfig()
	t, err := token.SignedString([]byte(config.JwtToken))
	return LoginInfo{
		AccessToken: t,
		Role:        role,
	}, err
}

func (s *Service) Create(username, password string, role entity.Role) error {
	if err := validateUsername(&username); err != nil {
		return err
	}
	if err := validatePassword(&password); err != nil {
		return err
	}
	if err := validateRole(role); err != nil {
		return err
	}

	salt, err := getSalt()
	if err != nil {
		return err
	}
	hash, err := getHash(password, salt)
	if err != nil {
		return err
	}

	r, err := db.Exec("INSERT INTO user (username, password, salt, role) VALUES (?, ?, ?, ?)", username, hash, salt, role)
	if err != nil {
		return err
	}
	id, _ := r.LastInsertId()

	// var users []entity.User
	// sqlscan.Select(context.Background(), conn, &users, "SELECT id, username, password, salt, role FROM user")
	// fmt.Println(users)
	log.Printf("Created user '%d' with Username %s", id, username)
	return nil
}

func (s *Service) UpdateRole(id int64, role entity.Role) error {
	if err := validateRole(role); err != nil {
		return err
	}
	if _, err := db.Exec("UPDATE user SET role=$1 WHERE id=$2", role, id); err != nil {
		return err
	}
	log.Printf("Updated role of user '%d'\n", id)
	return nil
}

func (s *Service) UpdatePassword(id int64, password string) error {
	if err := validatePassword(&password); err != nil {
		return err
	}

	var salt *string
	if err := db.Get(&salt, "SELECT salt FROM user WHERE id=?", id); err != nil {
		return err
	}
	if salt != nil {
		var err error
		password, err = getHash(password, *salt)
		if err != nil {
			return err
		}
	}
	if _, err := db.Exec("UPDATE user SET password=$1 WHERE id=$2", password, id); err != nil {
		return err
	}
	log.Printf("Updated password of user '%d'\n", id)
	return nil
}

func (s *Service) Delete(id int64) error {
	r, err := db.Exec("DELETE FROM user WHERE id=?", id)
	if err != nil {
		return err
	}
	rows, _ := r.RowsAffected()
	if rows < 1 {
		return fmt.Errorf("User '%d' does not exist", id)
	}
	log.Printf("Deleted user '%d'", id)
	return nil
}

func validateUsername(username *string) error {
	u := strings.TrimSpace(strings.ToLower(*username))
	if len(u) < minUsernameLength {
		return errors.New("Invalid username length")
	}
	*username = u
	return nil
}

func validatePassword(password *string) error {
	p := strings.TrimSpace(strings.ToLower(*password))
	if len(p) < minPasswordLength {
		return errors.New("Invalid password length")
	}
	*password = p
	return nil
}

func validateRole(role entity.Role) error {
	if role == entity.RoleNone {
		return errors.New("Invalid role")
	}
	return nil
}

func getHash(password, salt string) (string, error) {
	sb, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return "", fmt.Errorf("Unable to decode salt: %w", err)
	}
	hash := pbkdf2.Key([]byte(password), sb, iterations, hashSize, sha1.New)
	return base64.StdEncoding.EncodeToString(hash), nil
}

func getSalt() (string, error) {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("Unable to generate salt: %w", err)
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}
