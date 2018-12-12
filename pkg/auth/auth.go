package auth

import (
	"crypto/ecdsa"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"net/http"
	"time"
)

type jwtSigner struct {
	privKey *ecdsa.PrivateKey
}

func NewEcJwtSigner(privateKey []byte) (Signer, error) {
	privKey, err := jwt.ParseECPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil, err
	}

	return &jwtSigner{
		privKey: privKey,
	}, nil
}

func (j *jwtSigner) Sign(userDetails UserDetails) string {
	token := jwt.NewWithClaims(jwt.SigningMethodES512, jwt.MapClaims{
		"sub": userDetails.Id,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(j.privKey)
	if err != nil {
		panic(err)
	}

	return tokenString
}

type jwtAuthenticator struct {
	pkey *ecdsa.PublicKey
}

func NewEcJwtAuthenticator(validatorPublicKey []byte) (HttpRequestAuthenticator, error) {
	pkey, err := jwt.ParseECPublicKeyFromPEM(validatorPublicKey)
	if err != nil {
		return nil, err
	}

	return &jwtAuthenticator{
		pkey: pkey,
	}, nil
}

func (j *jwtAuthenticator) Authenticate(r *http.Request) *UserDetails {
	cookie, err := r.Cookie(loginCookieName)
	if err == http.ErrNoCookie {
		return nil
	}

	claims := j.getValidatedClaims(cookie.Value)
	if claims == nil {
		return nil
	}

	return &UserDetails{
		Id: claims["sub"].(string),
	}
}

func (j *jwtAuthenticator) getValidatedClaims(jwtString string) jwt.MapClaims {
	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return j.pkey, nil
	})

	if err != nil {
		return nil
	}

	claims := token.Claims.(jwt.MapClaims)
	return claims
}
