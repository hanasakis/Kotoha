package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig
	DB        DBConfig
	Redis     RedisConfig
	Milvus    MilvusConfig
	S3        S3Config
	Stripe    StripeConfig
	JWT       JWTConfig
	Langfuse  LangfuseConfig
	Ollama    OllamaConfig
	LLM       LLMConfig
	Embedding EmbeddingConfig
	I18n      I18nConfig
	CORS      CORSConfig
}

type ServerConfig struct {
	Host, Port, Env, LogLevel string
}
type DBConfig struct {
	Host, Port, User, Password, DBName, SSLMode string
	MaxOpenConns, MaxIdleConns                   int
}
type RedisConfig struct {
	Host, Port, Password string
	DB                   int
}
type MilvusConfig struct {
	Host, Port, User, Password, DB              string
	HNSWM, HNSWEfConstruct, EfSearch, RecallTopK int
	DenseWeight, BM25Weight                       float64
}
type S3Config struct {
	Endpoint, AccessKey, SecretKey, Bucket string
	UseSSL                                  bool
}
type StripeConfig struct {
	SecretKey, PublishableKey, WebhookSecret, Currency string
}
type JWTConfig struct {
	Secret, AccessTTL, RefreshTTL string
}
type LangfuseConfig struct {
	PublicKey, SecretKey, Host string
}
type OllamaConfig struct {
	Host string
}
type LLMConfig struct {
	Model       string
	MaxTokens   int
	Temperature float64
}
type EmbeddingConfig struct {
	Model string
	Dim   int
}
type I18nConfig struct {
	DefaultLocale    string
	SupportedLocales []string
}
type CORSConfig struct {
	Origins []string
}

func Load() (*Config, error) {
	envPath := findEnvFile()
	if envPath != "" {
		viper.SetConfigFile(envPath)
	} else {
		viper.SetConfigFile(".env")
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	_ = viper.ReadInConfig()

	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"), Port: getEnv("SERVER_PORT", "8080"),
			Env: getEnv("ENV", "development"), LogLevel: getEnv("LOG_LEVEL", "debug"),
		},
		DB: DBConfig{
			Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
			User: getEnv("DB_USER", "kotoha"), Password: getEnv("DB_PASSWORD", ""),
			DBName: getEnv("DB_NAME", "kotoha"), SSLMode: getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25), MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 10),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"), Port: getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""), DB: getEnvInt("REDIS_DB", 0),
		},
		Milvus: MilvusConfig{
			Host: getEnv("MILVUS_HOST", "localhost"), Port: getEnv("MILVUS_PORT", "19530"),
			User: getEnv("MILVUS_USER", "root"), Password: getEnv("MILVUS_PASSWORD", ""),
			DB: getEnv("MILVUS_DB", "kotoha"),
			HNSWM: getEnvInt("MILVUS_HNSW_M", 30), HNSWEfConstruct: getEnvInt("MILVUS_HNSW_EF_CONSTRUCTION", 64),
			EfSearch: getEnvInt("MILVUS_EF_SEARCH", 80),
			DenseWeight: getEnvFloat("MILVUS_DENSE_WEIGHT", 0.65), BM25Weight: getEnvFloat("MILVUS_BM25_WEIGHT", 0.35),
			RecallTopK: getEnvInt("MILVUS_RECALL_TOP_K", 40),
		},
		S3: S3Config{
			Endpoint: getEnv("S3_ENDPOINT", "localhost:9000"), AccessKey: getEnv("S3_ACCESS_KEY", ""),
			SecretKey: getEnv("S3_SECRET_KEY", ""), Bucket: getEnv("S3_BUCKET", "kotoha-images"),
			UseSSL: getEnvBool("S3_USE_SSL", false),
		},
		Stripe: StripeConfig{
			SecretKey: getEnv("STRIPE_SECRET_KEY", ""), PublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
			WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""), Currency: getEnv("STRIPE_CURRENCY", "cny"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""), AccessTTL: getEnv("JWT_ACCESS_TTL", "15m"),
			RefreshTTL: getEnv("JWT_REFRESH_TTL", "720h"),
		},
		Langfuse: LangfuseConfig{
			PublicKey: getEnv("LANGFUSE_PUBLIC_KEY", ""), SecretKey: getEnv("LANGFUSE_SECRET_KEY", ""),
			Host: getEnv("LANGFUSE_HOST", "https://cloud.langfuse.com"),
		},
		Ollama: OllamaConfig{
			Host: getEnv("OLLAMA_HOST", "http://localhost:11434"),
		},
		LLM: LLMConfig{
			Model: getEnv("LLM_MODEL", "qwen3.5:9b"), MaxTokens: getEnvInt("LLM_MAX_TOKENS", 2048),
			Temperature: getEnvFloat("LLM_TEMPERATURE", 0.3),
		},
		Embedding: EmbeddingConfig{
			Model: getEnv("EMBEDDING_MODEL", "bge-large-zh-v1.5"), Dim: getEnvInt("EMBEDDING_DIM", 1024),
		},
		I18n: I18nConfig{
			DefaultLocale: getEnv("DEFAULT_LOCALE", "zh"),
			SupportedLocales: strings.Split(getEnv("SUPPORTED_LOCALES", "zh,en"), ","),
		},
		CORS: CORSConfig{
			Origins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ","),
		},
	}, nil
}

func (c *Config) DSN() string {
	return "host=" + c.DB.Host + " user=" + c.DB.User + " password=" + c.DB.Password +
		" dbname=" + c.DB.DBName + " port=" + c.DB.Port + " sslmode=" + c.DB.SSLMode
}

func (c *Config) RedisAddr() string  { return c.Redis.Host + ":" + c.Redis.Port }
func (c *Config) MilvusAddr() string { return c.Milvus.Host + ":" + c.Milvus.Port }

func findEnvFile() string {
	cwd, _ := os.Getwd()
	for range 10 {
		path := filepath.Join(cwd, ".env")
		if _, err := os.Stat(path); err == nil {
			return path
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}
	return ""
}

func getEnv(key, def string) string {
	if v := viper.GetString(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := viper.GetInt(key); v != 0 || viper.IsSet(key) {
		return v
	}
	return def
}

func getEnvFloat(key string, def float64) float64 {
	if v := viper.GetFloat64(key); v != 0 || viper.IsSet(key) {
		return v
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if viper.IsSet(key) {
		return viper.GetBool(key)
	}
	return def
}
