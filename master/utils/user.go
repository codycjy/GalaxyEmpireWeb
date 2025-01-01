package utils

import (
	"context"
)

func UserIDFromContext(ctx context.Context) uint {
	userID, ok := ctx.Value("userID").(uint)
	if !ok {
		return 0
	}
	return userID
}

func GetRoleFromContext(ctx context.Context) int {
	role, exists := ctx.Value("role").(int)
	if !exists {
		return 0
	}
	return role
}
