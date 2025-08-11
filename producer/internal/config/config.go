package config

import "os"

type Config struct {
	Port          string
	KafkaBroker   string
	RedisAddr     string
	MongoURI      string
	MongoDatabase string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Load() Config {
	return Config{
		Port:          getEnv("PORT", "8081"),
		KafkaBroker:   getEnv("KAFKA_BROKER", "localhost:9094"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		MongoURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGODB_DATABASE", "restaurantdb"),
	}
}
