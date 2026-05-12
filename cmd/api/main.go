package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"profile-backend/internal/ai"
	"profile-backend/internal/contact"
	"profile-backend/internal/dashboard"
	"profile-backend/internal/notify"
	"profile-backend/internal/personal"
	"profile-backend/internal/profile"
	"profile-backend/internal/storage"
	"profile-backend/internal/web"
)

func main() {
	port := env("PORT", "8080")
	adminToken := os.Getenv("ADMIN_TOKEN")
	allowedOrigins := splitCSV(env("ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000"))
	storageDriver := strings.ToLower(env("STORAGE_DRIVER", "file"))

	var contactStore contact.Store
	var profileStore profile.Store
	var personalStore personal.Store

	switch storageDriver {
	case "postgres", "postgresql", "neon":
		pool, err := storage.ConnectPostgres(context.Background(), storage.PostgresConfig{
			DatabaseURL: env("DATABASE_URL", ""),
		})
		if err != nil {
			log.Fatalf("postgres connection failed: %v", err)
		}
		if env("AUTO_MIGRATE", "true") == "true" {
			if err := storage.MigratePostgres(context.Background(), pool); err != nil {
				log.Fatalf("postgres migration failed: %v", err)
			}
		}
		defer pool.Close()
		contactStore = contact.NewPostgresStore(pool)
		profileStore = profile.NewPostgresStore(pool)
		personalStore = personal.NewPostgresStore(pool)
	case "mongo", "mongodb":
		client, database, err := storage.ConnectMongo(context.Background(), storage.MongoConfig{
			URI:      env("MONGO_URI", "mongodb://localhost:27017"),
			Database: env("MONGO_DATABASE", "profile_os"),
		})
		if err != nil {
			log.Fatalf("mongo connection failed: %v", err)
		}
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := client.Disconnect(shutdownCtx); err != nil {
				log.Printf("mongo disconnect failed: %v", err)
			}
		}()
		contactStore = contact.NewMongoStore(database)
		profileStore = profile.NewMongoStore(database)
		personalStore = personal.NewMongoStore(database)
	case "file":
		contactStore = contact.NewFileStore(env("CONTACT_STORE_PATH", "data/contact_messages.jsonl"))
		profileStore = profile.NewFileStore(env("PROFILE_STORE_PATH", "data/profile.json"))
		personalStore = personal.NewFileStore(env("PERSONAL_STORE_PATH", "data/personal.json"))
	default:
		log.Fatalf("unsupported STORAGE_DRIVER %q; use postgres, mongo, or file", storageDriver)
	}

	contactHandler := contact.NewHandler(contactStore, adminToken)
	profileHandler := profile.NewHandler(profileStore, adminToken)
	dashboardHandler := dashboard.NewHandler(contactStore, profileStore, adminToken)
	personalHandler := personal.NewHandler(personalStore, adminToken)
	aiHandler := ai.NewHandler(personalStore, ai.NewClient(env("AI_PROVIDER", "local"), env("AI_API_KEY", ""), env("AI_MODEL", "")), adminToken)
	notifyHandler := notify.NewHandler(adminToken, env("FCM_SERVER_KEY", ""))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		web.WriteJSON(w, http.StatusOK, map[string]string{"service": "profile-backend", "status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
	})
	mux.HandleFunc("GET /api/profile", profileHandler.Get)
	mux.HandleFunc("PUT /api/profile", profileHandler.Update)
	mux.HandleFunc("POST /api/contact", contactHandler.Create)
	mux.HandleFunc("GET /api/contact/messages", contactHandler.List)
	mux.HandleFunc("PATCH /api/contact/messages/{id}", contactHandler.UpdateStatus)
	mux.HandleFunc("DELETE /api/contact/messages/{id}", contactHandler.Delete)
	mux.HandleFunc("GET /api/dashboard", dashboardHandler.Get)
	mux.HandleFunc("GET /api/personal/summary", personalHandler.Summary)
	mux.HandleFunc("GET /api/personal/notes", personalHandler.ListNotes)
	mux.HandleFunc("POST /api/personal/notes", personalHandler.CreateNote)
	mux.HandleFunc("DELETE /api/personal/notes/{id}", personalHandler.DeleteNote)
	mux.HandleFunc("GET /api/personal/reminders", personalHandler.ListReminders)
	mux.HandleFunc("POST /api/personal/reminders", personalHandler.CreateReminder)
	mux.HandleFunc("PATCH /api/personal/reminders/{id}/toggle", personalHandler.ToggleReminder)
	mux.HandleFunc("DELETE /api/personal/reminders/{id}", personalHandler.DeleteReminder)
	mux.HandleFunc("GET /api/personal/transactions", personalHandler.ListTransactions)
	mux.HandleFunc("POST /api/personal/transactions", personalHandler.CreateTransaction)
	mux.HandleFunc("DELETE /api/personal/transactions/{id}", personalHandler.DeleteTransaction)
	mux.HandleFunc("GET /api/personal/habits", personalHandler.ListHabits)
	mux.HandleFunc("POST /api/personal/habits", personalHandler.CreateHabit)
	mux.HandleFunc("PATCH /api/personal/habits/{id}/check-in", personalHandler.CheckInHabit)
	mux.HandleFunc("DELETE /api/personal/habits/{id}", personalHandler.DeleteHabit)
	mux.HandleFunc("POST /api/ai/daily-briefing", aiHandler.DailyBriefing)
	mux.HandleFunc("POST /api/ai/note-summary", aiHandler.NoteSummary)
	mux.HandleFunc("GET /api/notifications/status", notifyHandler.Status)
	mux.HandleFunc("POST /api/notifications/test", notifyHandler.SendTest)

	server := &http.Server{Addr: ":" + port, Handler: web.WithCORS(allowedOrigins, web.WithLogging(mux)), ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("Profile Backend listening on http://localhost:%s using %s storage", port, storageDriver)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
