package test

import (
	"crypto/x509"
	"encoding/pem"

	"github.com/golang-jwt/jwt/v5"
)

func NewToken(claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	block, _ := pem.Decode([]byte(RS256PrivateKey))
	if block == nil {
		panic("can't decpde private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		panic(err)
	}

	return signedToken
}
