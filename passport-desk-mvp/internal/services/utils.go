package services

import "context"

func getOperatorIDFromContext(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	if v := ctx.Value("user_id"); v != nil {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}
