package main

import (
	"context"
	"log"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/repository"
	"cerdasind-backend/pkg/database"
	"cerdasind-backend/pkg/utils"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db := database.InitDB()
	defer db.Close()

	userRepo := repository.NewUserRepository(db)

	username := "admin"
	email := "admin@cerdasind.com"
	password := "admin123"

	// Check if exists
	existing, _ := userRepo.FindByEmail(context.Background(), email)
	if existing != nil {
		log.Println("Admin user already exists")
		return
	}

	hash, _ := utils.HashPassword(password)
	admin := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Role:         model.RoleAdmin,
	}

	err := userRepo.Create(context.Background(), admin)
	if err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}

	log.Printf("Admin account created! Username: %s, Password: %s\n", username, password)
}
