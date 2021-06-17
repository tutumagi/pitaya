package metapart

import uuid "github.com/satori/go.uuid"

func NewUUID() string {
	return uuid.NewV1().String()
}
