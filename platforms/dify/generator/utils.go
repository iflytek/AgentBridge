package generator

import (
	"crypto/rand"
	"fmt"
	"time"
)

// generateRandomUUID generates a random UUID v4.
func generateRandomUUID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

// generateShortID generates a short random hexadecimal string.
func generateShortID(length int) string {
	bytes := make([]byte, (length+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano()%10000)
	}
	
	result := fmt.Sprintf("%x", bytes)
	if len(result) > length {
		result = result[:length]
	}
	
	return result
}
