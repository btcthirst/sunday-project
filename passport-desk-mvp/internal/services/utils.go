package services

import "context"

func getOperatorIDFromContext(ctx context.Context) int64 {
	return ctx.Value("user_id").(int64)
}
