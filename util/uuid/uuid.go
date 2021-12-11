package uuid

import "github.com/google/uuid"

func StrUUID() string {
	u, _ := uuid.NewUUID()
	return u.String()
}
