package utils

import "context"

func UserIDFromContext(ctx context.Context) uint {
	userID, err := ctx.Value("userID").(uint)
	if !err {
		return 0
	}
	return userID
}
