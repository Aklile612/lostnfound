package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	UploadDir      string
	CORSOrigin     string
	GeminiAPIKey   string
	GroqAPIKey     string
	GeminiModel    string
	GroqModel      string
	MatchWindow    int
	MatchThreshold float64
}

func Load() (Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	host := env("POSTGRES_HOST", "localhost")
	port := env("POSTGRES_PORT", "5432")
	user := env("POSTGRES_USER", "nemma")
	pass := env("POSTGRES_PASSWORD", "nemma")
	name := env("POSTGRES_DB", "lost_and_found")
	dbURL := env("DATABASE_URL", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name))

	window, err := strconv.Atoi(env("MATCH_WINDOW_DAYS", "7"))
	if err != nil {
		window = 7
	}
	threshold, err := strconv.ParseFloat(env("MATCH_THRESHOLD", "40"), 64)
	if err != nil {
		threshold = 40
	}

	cfg := Config{
		Port:           env("PORT", "8080"),
		DatabaseURL:    dbURL,
		UploadDir:      env("UPLOAD_DIR", "./uploads"),
		CORSOrigin:     env("CORS_ORIGIN", "http://localhost:3000"),
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
		GroqAPIKey:     os.Getenv("GROQ_API_KEY"),
		GeminiModel:    env("GEMINI_MODEL", "gemini-3.6-flash"),
		GroqModel:      env("GROQ_MODEL", "openai/gpt-oss-20b"),
		MatchWindow:    window,
		MatchThreshold: threshold,
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
