package service

import "context"

var _ Service = &MockService{}

type MockService struct {
	GenerateTokenFunc func(ctx context.Context, creds Credentials) (*Token, error)
	VerifyTokenFunc   func(ctx context.Context, token string) error
	SumFunc           func(ctx context.Context, data any) (string, error)
}

func (m *MockService) GenerateToken(ctx context.Context, creds Credentials) (*Token, error) {
	return m.GenerateTokenFunc(ctx, creds)
}

func (m *MockService) VerifyToken(ctx context.Context, token string) error {
	return m.VerifyTokenFunc(ctx, token)
}

func (m *MockService) Sum(ctx context.Context, data any) (string, error) {
	return m.SumFunc(ctx, data)
}
