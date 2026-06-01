package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/encrypt"
)

func DecryptBody(cipher *encrypt.Cipher) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cipher == nil {
			c.Next()
			return
		}

		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodDelete || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		isEncrypted := contentType == "application/x-encrypted" || c.GetHeader("X-Encrypted") == "true"

		if !isEncrypted {
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_code": "INVALID_ENCRYPTED_BODY",
				"message":    "failed to read request body",
			})
			return
		}
		c.Request.Body.Close()

		plaintext, err := cipher.DecryptFromHex(string(body))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_code": "DECRYPTION_FAILED",
				"message":    "failed to decrypt request body",
			})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBufferString(plaintext))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.ContentLength = int64(len(plaintext))

		c.Next()
	}
}
