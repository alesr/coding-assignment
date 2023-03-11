package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/alesr/code-assignment/internal/service"
	"github.com/go-chi/chi"
)

const bearerPrefix = "Bearer "

// RESTApp is the REST server.
type RESTApp struct {
	logger     *zap.Logger
	httpServer *http.Server
	svc        service.Service
}

// NewRESTApp creates a new RESTApp instance with configured routes.
func NewRESTApp(logger *zap.Logger, port string, router chi.Router, svc service.Service) *RESTApp {
	app := RESTApp{
		logger: logger,
		svc:    svc,
	}

	router.Post("/auth", app.authHandler)
	router.Post("/sum", app.sumHandler)

	app.httpServer = &http.Server{
		Handler: router,
		Addr:    net.JoinHostPort("", port),
	}
	return &app
}

// Start starts the REST server.
func (r *RESTApp) Start() error {
	r.logger.Info("starting REST app on " + r.httpServer.Addr)

	if err := r.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("could not start REST app: %w", err)
	}
	return nil
}

// Stop gracefully stops the REST server.
func (r *RESTApp) Stop(ctx context.Context) error {
	r.logger.Info("stopping REST app")

	if err := r.httpServer.Shutdown(ctx); err != nil {
		r.logger.Error("failed to shutdown REST app", zap.Error(err))
		return err
	}
	return nil
}

// HTTP handlers.

func (app *RESTApp) authHandler(w http.ResponseWriter, r *http.Request) {
	var authReq authenticateRequest
	if err := json.NewDecoder(r.Body).Decode(&authReq); err != nil {
		app.logger.Error("could not decode request", zap.Error(err))
		writeJSONError(w, ErrInvalidRequest)
		return
	}

	if err := authReq.validate(); err != nil {
		app.logger.Error("could not validate request", zap.Error(err))
		writeJSONError(w, err)
		return
	}

	creds := service.Credentials{
		Username: authReq.Username,
		Password: authReq.Password,
	}

	token, err := app.svc.GenerateToken(r.Context(), creds)
	if err != nil {
		app.logger.Error("could not generate token", zap.Error(err))
		writeJSONError(w, toTransportError(err))
		return
	}

	writeJSON(w, authenticaResponse{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		ExpiresIn:   token.ExpiresIn,
	})
}

func (app *RESTApp) sumHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := extractTokenFromHeader(r.Header.Get("Authorization"))
	if tokenString == "" {
		app.logger.Warn("missing token")
		writeJSONError(w, ErrUnauthorized)
		return
	}

	var sumReq sumRequest
	if err := json.NewDecoder(r.Body).Decode(&sumReq); err != nil {
		app.logger.Error("could not decode request", zap.Error(err))
		writeJSONError(w, ErrInvalidRequest)
		return
	}

	if err := app.svc.VerifyToken(r.Context(), tokenString); err != nil {
		app.logger.Warn("could not verify token", zap.Error(err))
		writeJSONError(w, ErrUnauthorized)
		return
	}

	sum, err := app.svc.Sum(r.Context(), sumReq)
	if err != nil {
		app.logger.Error("could not sum", zap.Error(err))
		writeJSONError(w, toTransportError(err))
		return
	}
	writeJSON(w, sumResponse{Sum: sum})
}

func extractTokenFromHeader(authHeader string) string {
	if strings.HasPrefix(authHeader, bearerPrefix) {
		return authHeader[len(bearerPrefix):]
	}
	return ""
}
