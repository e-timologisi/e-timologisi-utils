package crypto_utils

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"

	"github.com/dgrijalva/jwt-go"

	"github.com/e-timologisi/e-timologisi-utils/http_utils"
	"github.com/e-timologisi/e-timologisi-utils/logger"
	"github.com/e-timologisi/e-timologisi-utils/resterrors"
)

var (
	KeyGenerator = []byte("NcRfUjXn2r5u8x/A")
	MySigningKey = []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJjbGllbnQiOiJKaXNjIiwiZXhwIjowfQ.CYwly1IRD_F9ptpSx8OJQVL2jgwMpkxlR5HheQLF9us")
)

func GenerateJWT(username, userId string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = username
	claims["exp"] = 0
	claims["user_id"] = userId
	tokenString, err := token.SignedString(MySigningKey)

	if err != nil {

		return "", fmt.Errorf("Something Went Wrong: %s", err.Error())
	}

	return tokenString, nil
}

func EncryptPass(key []byte, text string) string {
	// key := []byte(keyText)
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Error("", err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		logger.Error("", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// decrypt from base64 to decrypted string
func DecryptPass(key []byte, cryptoText string) string {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Error("", err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		logger.Error("ciphertext too short", err)
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext)
}
func ValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			authErr := resterrors.NewUnauthorizedError("Unauthorized")
			http_utils.RespondWithError(w, authErr.Status(), authErr)
			return
		}

		tokenString := strings.Split(authorizationHeader, " ")[0]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the token signing method is as expected
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Unexpected signing method")
			}

			// Provide the key used to sign the token
			return []byte(MySigningKey), nil
		})

		if err != nil {
			authErr := resterrors.NewUnauthorizedError("Unauthorized")
			http_utils.RespondWithError(w, authErr.Status(), authErr)
			return
		}

		if !token.Valid {
			authErr := resterrors.NewUnauthorizedError("Unauthorized")
			http_utils.RespondWithError(w, authErr.Status(), authErr)
			return
		}

		// Add user_id from the token to the request context for later use
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			authErr := resterrors.NewUnauthorizedError("Unauthorized")
			http_utils.RespondWithError(w, authErr.Status(), authErr)
			return
		}
		userID, ok := claims["user_id"].(string)
		if !ok {
			authErr := resterrors.NewUnauthorizedError("Unauthorized")
			http_utils.RespondWithError(w, authErr.Status(), authErr)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
