package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
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
	var privKey *rsa.PrivateKey

	if privateKeyPath != "" {
		pemBytes, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read RSA private key file: %w", err)
		}

		block, _ := pem.Decode(pemBytes)
		if block == nil {
			return nil, errors.New("failed to decode PEM block from private key file")
		}

		// Support both PKCS#1 and PKCS#8
		if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
			privKey = k
		} else if keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			var ok bool
			privKey, ok = keyInterface.(*rsa.PrivateKey)
			if !ok {
				return nil, errors.New("PKCS#8 key is not an RSA private key")
			}
		} else {
			return nil, fmt.Errorf("could not parse private key as PKCS#1 or PKCS#8: %w", err)
		}
	} else {
		// Development fallback: auto-generate in-memory
		var err error
		privKey, err = rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, fmt.Errorf("failed to generate fallback RSA key: %w", err)
		}
	}

	pubKey := &privKey.PublicKey
	// Deterministic RFC 7638 Key ID based on SHA-256 of the public modulus
	hash := sha256.Sum256(pubKey.N.Bytes())
	kid := hex.EncodeToString(hash[:8])

	return &KeyManager{
		PrivateKey: privKey,
		PublicKey:  pubKey,
		KeyID:      kid,
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