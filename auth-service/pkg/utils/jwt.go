package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"time"

	"auth-service/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type KeyManager struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	KeyID      string
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

func NewKeyManager(privateKeyPath string) (*KeyManager, error) {
	if privateKeyPath != "" {
		pemBytes, err := os.ReadFile(privateKeyPath)
		if err == nil {
			block, _ := pem.Decode(pemBytes)
			if block != nil {
				key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
				if err == nil {
					return &KeyManager{
						PrivateKey: key,
						PublicKey:  &key.PublicKey,
						KeyID:      "auth-service-key-1",
					}, nil
				}
			}
		}
	}

	// Auto-generate cryptographically secure RSA-4096 key pair if no file provided
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	return &KeyManager{
		PrivateKey: privKey,
		PublicKey:  &privKey.PublicKey,
		KeyID:      "auth-service-key-primary",
	}, nil
}

func (km *KeyManager) GenerateAccessToken(userID uuid.UUID, email string, role domain.Role, ttl time.Duration) (string, error) {
	claims := &domain.JwtCustomClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    "auth-service",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = km.KeyID

	return token.SignedString(km.PrivateKey)
}

func (km *KeyManager) ValidateAccessToken(tokenString string) (*domain.JwtCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method: expected RS256")
		}
		return km.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*domain.JwtCustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

func (km *KeyManager) GetJWKS() JWKS {
	nBytes := km.PublicKey.N.Bytes()
	eBytes := big.NewInt(int64(km.PublicKey.E)).Bytes()

	return JWKS{
		Keys: []JWK{
			{
				Kty: "RSA",
				Kid: km.KeyID,
				Use: "sig",
				Alg: "RS256",
				N:   base64.RawURLEncoding.EncodeToString(nBytes),
				E:   base64.RawURLEncoding.EncodeToString(eBytes),
			},
		},
	}
}

func GenerateSecureToken(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}