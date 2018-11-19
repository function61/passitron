package auth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"net/http"
	"time"
)

type UserDetails struct {
	UserId string
}

type HttpRequestAuthenticator interface {
	Authenticate(*http.Request) *UserDetails
}

type Signer interface {
	Sign(details UserDetails) string
}

type jwtSigner struct {
	privKey *ecdsa.PrivateKey
}

func NewJwtSigner(privateKey []byte) (Signer, error) {
	privKey, err := jwt.ParseECPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil, err
	}

	return &jwtSigner{
		privKey: privKey,
	}, nil
}

func (j *jwtSigner) Sign(details UserDetails) string {
	token := jwt.NewWithClaims(jwt.SigningMethodES512, jwt.MapClaims{
		"sub": details.UserId,
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

func NewJwtAuthenticator(validatorPublicKey []byte) (HttpRequestAuthenticator, error) {
	pkey, err := jwt.ParseECPublicKeyFromPEM(validatorPublicKey)
	if err != nil {
		return nil, err
	}

	return &jwtAuthenticator{
		pkey: pkey,
	}, nil
}

func (j *jwtAuthenticator) Authenticate(r *http.Request) *UserDetails {
	cookie, err := r.Cookie("login")
	if err == http.ErrNoCookie {
		return nil
	}

	claims := j.getValidatedClaims(cookie.Value)
	if claims == nil {
		return nil
	}

	return &UserDetails{
		UserId: claims["sub"].(string),
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

func ToCookie(tokenString string) *http.Cookie {
	return &http.Cookie{
		Name:     "login",
		Path:     "/",
		Value:    tokenString,
		HttpOnly: true, // = not visible to JavaScript
		// Secure: true, // FIXME
	}
}

func GenerateKey() ([]byte, []byte, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	marshaledPrivKey, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return nil, nil, err
	}
	marshaledPubKey, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	pemPrivKey := &bytes.Buffer{}
	pemPubKey := &bytes.Buffer{}

	if err := pem.Encode(pemPrivKey, &pem.Block{Type: "PRIVATE KEY", Bytes: marshaledPrivKey}); err != nil {
		return nil, nil, err
	}

	if err := pem.Encode(pemPubKey, &pem.Block{Type: "PUBLIC KEY", Bytes: marshaledPubKey}); err != nil {
		return nil, nil, err
	}

	return pemPrivKey.Bytes(), pemPubKey.Bytes(), nil
}

/*
func DeleteLoginCookie() *http.Cookie {
	return &http.Cookie{
		Name:     "login",
		Value:    "del",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // => delete
		// Secure: true, // FIXME
	}
}
*/
