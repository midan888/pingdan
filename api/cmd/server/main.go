package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/pingdan/api/internal/alerts"
	"github.com/pingdan/api/internal/assertions"
	"github.com/pingdan/api/internal/auth"
	"github.com/pingdan/api/internal/checks"
	"github.com/pingdan/api/internal/config"
	"github.com/pingdan/api/internal/db"
	"github.com/pingdan/api/internal/endpoints"
	"github.com/pingdan/api/internal/groups"
	httpx "github.com/pingdan/api/internal/http"
	"github.com/pingdan/api/internal/pinger"
	"github.com/pingdan/api/internal/sslcheck"
	"github.com/pingdan/api/internal/statuspages"
)

func main() {
	_ = godotenv.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		logger.Error("migrate", "err", err)
		os.Exit(1)
	}

	jwt := auth.NewJWT(cfg.JWTSecret, cfg.JWTTTL)
	oauthSvc := auth.NewOAuth(pool, jwt, cfg.PublicURL,
		cfg.GoogleClientID, cfg.GoogleClientSecret,
		cfg.GitHubClientID, cfg.GitHubClientSecret,
	)
	emailSvc := auth.NewEmail(pool, jwt)

	endpointStore := &endpoints.Store{Pool: pool}
	groupStore := &groups.Store{Pool: pool}
	checkStore := &checks.Store{Pool: pool}
	assertionStore := &assertions.Store{Pool: pool}
	statusPageStore := &statuspages.Store{Pool: pool}
	dispatcher := &alerts.Dispatcher{
		Pool: pool, Logger: logger,
		ResendAPIKey:     cfg.ResendAPIKey,
		EmailFrom:        cfg.EmailFrom,
		TelegramBotToken: cfg.TelegramBotToken,
		PushoverAppToken: cfg.PushoverAppToken,
		TwilioAccountSID: cfg.TwilioAccountSID,
		TwilioAuthToken:  cfg.TwilioAuthToken,
		TwilioFrom:       cfg.TwilioFrom,
	}
	scheduler := pinger.NewScheduler(ctx, endpointStore, checkStore, assertionStore, dispatcher, logger)
	if err := scheduler.Start(ctx); err != nil {
		logger.Error("scheduler start", "err", err)
		os.Exit(1)
	}

	// Daily TLS-certificate expiry monitor for HTTPS endpoints.
	sslChecker := sslcheck.New(endpointStore, dispatcher, logger)
	go sslChecker.Run(ctx)

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })

	authH := &httpx.AuthHandlers{OAuth: oauthSvc, Email: emailSvc, JWT: jwt, FrontendURL: cfg.FrontendURL}
	authH.Routes(r)

	// Public, unauthenticated status pages.
	pubStatusH := &httpx.PublicStatusHandlers{Store: statusPageStore, Endpoints: endpointStore, Checks: checkStore}
	pubStatusH.Routes(r)

	r.Group(func(r chi.Router) {
		r.Use(httpx.AuthMiddleware(jwt))
		authH.AuthedRoutes(r)
		r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			u := httpx.UserFrom(r.Context())
			httpx.WriteJSON(w, 200, map[string]any{"id": u.ID, "email": u.Email, "isAdmin": cfg.IsAdmin(u.Email)})
		})
		epH := &httpx.EndpointHandlers{Store: endpointStore, Checks: checkStore, Assertions: assertionStore, Scheduler: scheduler, SSL: sslChecker, Pool: pool}
		epH.Routes(r)
		alH := &httpx.AlertHandlers{Pool: pool, Dispatcher: dispatcher}
		alH.Routes(r)
		grH := &httpx.GroupHandlers{Store: groupStore}
		grH.Routes(r)
		spH := &httpx.StatusPageHandlers{Store: statusPageStore, Endpoints: endpointStore}
		spH.Routes(r)
		adH := &httpx.AdminHandlers{Pool: pool, IsAdmin: cfg.IsAdmin}
		adH.Routes(r)
	})

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: r, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		logger.Info("listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http", "err", err)
			cancel()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	_ = srv.Shutdown(shutdownCtx)
}
