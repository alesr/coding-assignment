package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

const (
	jtwClaimDuration time.Duration = time.Hour
	jwtClaimIssuer   string        = "foo-issuer"
	jwtClaimAudience string        = "foo-audience"
	tokenType        string        = "Bearer"
)

type Claims struct {
	jwt.StandardClaims
}

var _ Service = &DefaultService{}

// DefaultService is the default implementation of the Service interface.
type DefaultService struct {
	logger *zap.Logger
	jwtKey []byte
}

// NewDefaultService creates a new DefaultService.
func NewDefaultService(logger *zap.Logger, jwtKey []byte) *DefaultService {
	return &DefaultService{
		logger: logger,
		jwtKey: jwtKey,
	}
}

// GenerateToken generates a JWT token for the provided credentials.
func (s *DefaultService) GenerateToken(ctx context.Context, creds Credentials) (*Token, error) {
	// Validate credentials
	if err := creds.validate(); err != nil {
		return nil, fmt.Errorf("could not invalid credentials: %w", err)
	}

	// Create claims with username as subject
	claims := Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(jtwClaimDuration).Unix(),
			Issuer:    jwtClaimIssuer,
			Audience:  jwtClaimAudience,
			Subject:   creds.Username,
			Id:        strconv.FormatInt(time.Now().Unix(), 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.jwtKey)
	if err != nil {
		return nil, fmt.Errorf("could not sign token: %w", err)
	}
	return &Token{
		AccessToken: signedToken,
		TokenType:   tokenType,
		ExpiresIn:   int64(jtwClaimDuration.Seconds()),
	}, nil
}

// VerifyToken verifies the provided JWT token.
func (s *DefaultService) VerifyToken(ctx context.Context, token string) error {
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		return s.jwtKey, nil
	})
	if err != nil {
		return fmt.Errorf("could not parse token: %w", err)
	}

	if !tkn.Valid {
		return ErrTokenInvalid
	}

	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return ErrTokenInvalidExpiration
	}

	if !claims.VerifyIssuer(jwtClaimIssuer, true) {
		return ErrTokenInvalidIssuer
	}

	if !claims.VerifyAudience(jwtClaimAudience, true) {
		return ErrTokenInvalidAudience
	}
	return nil
}

// Sum sums the provided data.
func (s *DefaultService) Sum(ctx context.Context, data any) (string, error) {
	result, err := sumNumbers(data)
	if err != nil {
		return "", fmt.Errorf("could not sum numbers: %w", err)
	}

	// Assuming that we don't log debug level in production.
	s.logger.Debug("generating hash for", zap.Float64("result", result))

	hash := sha256.Sum256([]byte(fmt.Sprintf("%f", result)))
	return fmt.Sprintf("%x", hash), nil
}

// sumNumbers sums the provided data.
// We could possible cover more cases but I think this is enough for the purpose of this exercise.
// It's also unliked that I wouldn't have clear requirements for this work.
func sumNumbers(data any) (float64, error) {
	switch val := data.(type) {

	case nil:
		return 0, nil

	case float64:
		return val, nil

	case int:
		return float64(val), nil

	case string:
		if val == "" {
			return 0, nil
		}

		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("could not parse string to float: %s,  %w", err, ErrUnsupportedValueType)
		}
		return num, nil

	case []float64:
		var sum float64
		for _, v := range val {
			sum += v
		}
		return sum, nil

	case []int:
		var sum float64
		for _, v := range val {
			sum += float64(v)
		}
		return sum, nil

	case []string:
		var sum float64
		for _, v := range val {
			if v == "" {
				continue
			}

			num, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return 0, fmt.Errorf("could not parse string to float: %s,  %w", err, ErrUnsupportedValueType)
			}
			sum += num
		}
		return sum, nil

	case []any:
		var sum float64
		for _, v := range val {
			result, err := sumNumbers(v)
			if err != nil {
				return 0, fmt.Errorf("could not sum numbers: %s,  %w", err, ErrUnsupportedValueType)
			}

			sum += result
		}
		return sum, nil

	case map[string]any:
		var sum float64
		for _, v := range val {
			result, err := sumNumbers(v)
			if err != nil {
				return 0, fmt.Errorf("could not sum numbers: %s,  %w", err, ErrUnsupportedValueType)
			}
			sum += result
		}
		return sum, nil

	default:
		return 0, ErrUnsupportedValueType
	}
}
