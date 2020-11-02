package controller

import (
	"net/http"

	"github.com/Erexo/Ventana/core/entity"
	"github.com/go-chi/jwtauth"
)

type Claims struct {
	UserId int64
	Hash   string
	Role   entity.Role
}

func ReadClaims(w http.ResponseWriter, r *http.Request) (Claims, bool) {
	token, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		Unauthorize(w)
		return Claims{}, false
	}
	if token == nil || !token.Valid {
		Unauthorize(w)
		return Claims{}, false
	}

	id, ok := claims["uid"].(float64)
	if !ok {
		Unauthorize(w)
		return Claims{}, false
	}
	hash, ok := claims["pwd"].(string)
	if !ok {
		Unauthorize(w)
		return Claims{}, false
	}
	role, ok := claims["role"].(float64)
	if !ok {
		Unauthorize(w)
		return Claims{}, false
	}
	return Claims{
		UserId: int64(id),
		Hash:   hash,
		Role:   entity.Role(role),
	}, true
}

func Unauthorize(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}
