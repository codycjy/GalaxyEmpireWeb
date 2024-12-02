package utils

import "context"

func UserIDFromContext(ctx context.Context) uint {
	userID, _ := ctx.Value("userId").(uint)
	return userID
}
