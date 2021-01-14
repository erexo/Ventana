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

	"github.com/Erexo/Ventana/core/domain"
	"github.com/Erexo/Ventana/core/dto"
	"github.com/Erexo/Ventana/core/entity"
	"github.com/Erexo/Ventana/infrastructure/config"
	"github.com/Erexo/Ventana/infrastructure/db"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/pbkdf2"
)

const (
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
	AccessToken string      `json:"accessToken"`
	Role        domain.Role `json:"role"`
}

func (s *Service) Login(username, password string) (LoginInfo, error) {
	conn, err := db.GetConnection()
	if err != nil {
		return LoginInfo{}, err
	}
	defer conn.Close()

	var data struct {
		Id       int64          `db:"id"`
		Password string         `db:"password"`
		Salt     sql.NullString `db:"salt"`
		Role     domain.Role    `db:"role"`
	}
	if err := db.Get(&data, "SELECT id, password, salt, role FROM user WHERE username LIKE ?", username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LoginInfo{}, ErrInvalidCredentials
		}
		return LoginInfo{}, err
	}

	if data.Salt.Valid {
		var err error
		password, err = getHash(password, data.Salt.String)
		if err != nil {
			return LoginInfo{}, err
		}
	}
	if password != data.Password {
		return LoginInfo{}, ErrInvalidCredentials
	}

	now := time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":  math.MaxInt32,
		"nbf":  now,
		"iat":  now,
		"sub":  username,
		"uid":  data.Id,
		"pwd":  data.Password,
		"role": data.Role,
	})
	config := config.GetConfig()
	t, err := token.SignedString([]byte(config.JwtToken))
	return LoginInfo{
		AccessToken: t,
		Role:        data.Role,
	}, err
}

func (s *Service) Browse(filters dto.Filters) ([]dto.User, error) {
	var ret []dto.User
	if err := filters.Validate(dto.User{}); err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT id, username, role FROM user%s", filters.GetQuery())

	err := db.Select(&ret, query)
	return ret, err
}

func (s *Service) Create(username, password string, role domain.Role) error {
	if err := entity.ValidateName(&username); err != nil {
		return fmt.Errorf("Username: %w", err)
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

	log.Printf("Created user '%d' with Username %s", id, username)
	return nil
}

func (s *Service) UpdateRole(id int64, role domain.Role) error {
	if err := validateRole(role); err != nil {
		return err
	}
	if _, err := db.Exec("UPDATE user SET role=? WHERE id=?", role, id); err != nil {
		return err
	}
	log.Printf("Updated role of user '%d'\n", id)
	return nil
}

func (s *Service) UpdatePassword(id int64, password string) error {
	if err := validatePassword(&password); err != nil {
		return err
	}

	var salt sql.NullString
	if err := db.Scan("SELECT salt FROM user WHERE id=?", id, &salt); err != nil {
		return err
	}
	if salt.Valid {
		var err error
		password, err = getHash(password, salt.String)
		if err != nil {
			return err
		}
	}

	if _, err := db.Exec("UPDATE user SET password=? WHERE id=?", password, id); err != nil {
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

func (s *Service) Verify(id int64, hash string, role domain.Role) bool {
	var data struct {
		Hash string      `db:"password"`
		Role domain.Role `db:"role"`
	}
	if err := db.Get(&data, "SELECT password, role FROM user WHERE id=?", id); err != nil {
		return false
	}
	return data.Hash == hash && data.Role == role
}

func validatePassword(password *string) error {
	p := strings.TrimSpace(strings.ToLower(*password))
	if len(p) < minPasswordLength {
		return errors.New("Invalid password length")
	}
	*password = p
	return nil
}

func validateRole(role domain.Role) error {
	if role == domain.RoleNone {
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
