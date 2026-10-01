package helper

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager mengelola pembuatan dan verifikasi JWT access token.
type JWTManager struct {
	secret []byte
	issuer string
	ttl    time.Duration // masa berlaku access token
}

// Claims adalah isi payload token kita.
// Kita simpan UserID dan Role agar middleware bisa cek tanpa query DB.
type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTManager membuat JWTManager baru.
func NewJWTManager(secret, issuer string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
	}
}

// Generate membuat JWT token baru untuk user tertentu.
// Token berisi user_id dan role sehingga middleware tidak perlu query DB.
func (m *JWTManager) Generate(userID int, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Verify memvalidasi token dan mengembalikan Claims-nya.
// Error dikembalikan kalau token expired, tanda tangan salah, atau format rusak.
func (m *JWTManager) Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
		func(t *jwt.Token) (any, error) {
			// Pastikan algoritma yang dipakai sesuai ekspektasi (HMAC)
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("metode signing tidak valid")
			}
			return m.secret, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token tidak valid")
	}

	return claims, nil
}

// TTLSeconds mengembalikan masa berlaku token dalam detik (untuk field expires_in).
func (m *JWTManager) TTLSeconds() int {
	return int(m.ttl.Seconds())
}
