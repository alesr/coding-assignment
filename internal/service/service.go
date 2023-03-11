package service

import "context"

type Service interface {
	GenerateToken(ctx context.Context, cred Credentials) (*Token, error)
	VerifyToken(ctx context.Context, token string) error
	Sum(ctx context.Context, data any) (string, error)
}
