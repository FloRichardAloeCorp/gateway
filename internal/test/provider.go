package test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func LaunchTestProvider() *httptest.Server {
	baseURL := ""

	router := gin.New()
	router.GET("/.well-known/openid-configuration", func(c *gin.Context) {
		type providerJSON struct {
			Issuer        string   `json:"issuer"`
			AuthURL       string   `json:"authorization_endpoint"`
			TokenURL      string   `json:"token_endpoint"`
			DeviceAuthURL string   `json:"device_authorization_endpoint"`
			JWKSURL       string   `json:"jwks_uri"`
			UserInfoURL   string   `json:"userinfo_endpoint"`
			Algorithms    []string `json:"id_token_signing_alg_values_supported"`
		}

		configuration := providerJSON{
			Issuer:        baseURL,
			AuthURL:       baseURL + "/auth",
			TokenURL:      baseURL + "/token",
			DeviceAuthURL: baseURL + "/device/auth",
			JWKSURL:       baseURL + "/jwks",
			UserInfoURL:   baseURL + "/user_info",
			Algorithms: []string{
				"RS256",
			},
		}

		c.JSON(http.StatusOK, configuration)
	})

	router.GET("/jwks", func(c *gin.Context) {
		type Key struct {
			Use string   `json:"use,omitempty"`
			Kty string   `json:"kty,omitempty"`
			Kid string   `json:"kid,omitempty"`
			Alg string   `json:"alg,omitempty"`
			E   string   `json:"e,omitempty"`
			N   string   `json:"n,omitempty"`
			X5c []string `json:"x5c,omitempty"`
		}

		type JWKSResponse struct {
			Keys []Key `json:"keys"`
		}

		block, _ := pem.Decode([]byte(RS256PublicKey))
		if block == nil {
			panic("can't decode public key")
		}

		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			panic("can't parse public key " + err.Error())
		}

		rsaPubKey, ok := pubKey.(*rsa.PublicKey)
		if !ok {
			panic("can't cast to RSA pub key")
		}

		response := JWKSResponse{
			Keys: []Key{
				{
					Use: "sig",
					Kid: "1",
					Kty: "RSA",
					Alg: "RS256",
					E:   "AQAB",
					N:   base64.RawURLEncoding.EncodeToString(rsaPubKey.N.Bytes()),
				},
			},
		}

		c.JSON(http.StatusOK, &response)
	})

	jwt.New(jwt.SigningMethodRS256)

	server := httptest.NewServer(router)
	baseURL = server.URL
	return server
}
