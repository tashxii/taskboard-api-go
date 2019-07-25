package common

import (
	"strings"

	"github.com/google/uuid"
)

// GenerateID returns unique id
func GenerateID() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}
