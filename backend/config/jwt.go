package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
)

var JWTSecret string

// InitializeJWTSecret initializes the JWT secret from environment or generates a secure random one
func InitializeJWTSecret() error {
	secret := os.Getenv("JWT_SECRET")
	env := os.Getenv("GO_ENV")

	// In production, JWT_SECRET is REQUIRED
	if env == "production" {
		if secret == "" || secret == "default_secret_key_change_me" {
			log.Fatal("[SECURITY ERROR] JWT_SECRET must be set in production environment!")
		}
		if len(secret) < 32 {
			log.Fatal("[SECURITY ERROR] JWT_SECRET must be at least 32 characters in production!")
		}
		JWTSecret = secret
		log.Println("[SECURITY] JWT_SECRET loaded from environment variable")
		return nil
	}

	// Check if JWT_SECRET is set and not the insecure default
	if secret != "" && secret != "default_secret_key_change_me" {
		JWTSecret = secret
		log.Println("[SECURITY] JWT_SECRET loaded from environment variable")
		return nil
	}

	// If JWT_SECRET is not set or is the insecure default, generate a random one
	if secret == "default_secret_key_change_me" {
		log.Println("[SECURITY WARNING] Detected insecure default JWT_SECRET")
	}

	// Generate a cryptographically secure random secret
	randomBytes := make([]byte, 32) // 256 bits
	_, err := rand.Read(randomBytes)
	if err != nil {
		return fmt.Errorf("failed to generate random JWT secret: %w", err)
	}

	JWTSecret = base64.URLEncoding.EncodeToString(randomBytes)

	log.Println("[SECURITY WARNING] JWT_SECRET not set in environment!")
	log.Println("[SECURITY WARNING] Generated random JWT_SECRET for this session")
	log.Println("[SECURITY WARNING] All tokens will be invalidated on server restart")
	log.Println("[SECURITY WARNING] Set JWT_SECRET environment variable for production!")

	return nil
}

// GetJWTSecret returns the current JWT secret
func GetJWTSecret() string {
	if JWTSecret == "" {
		log.Fatal("[SECURITY ERROR] JWT secret not initialized! Call InitializeJWTSecret() first")
	}
	return JWTSecret
}
