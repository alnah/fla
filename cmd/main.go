package main

import (
	"log"
	"time"

	"github.com/alnah/fla/internal/domain"
)

// RealClock implements the Clock interface using actual system time
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }

func main() {
	clock := RealClock{}

	// Create UserID
	userID, err := domain.NewUserID("test")
	if err != nil {
		log.Fatal("Failed to create UserID:", err)
	}

	// Create Username
	username, err := domain.NewUsername("alnah")
	if err != nil {
		log.Fatal("Failed to create Username:", err)
	}

	// Create Email
	email, err := domain.NewEmail("alexis.nahan@gmail.com")
	if err != nil {
		log.Fatal("Failed to create Email:", err)
	}

	// Create FirstName
	firstName, err := domain.NewFirstName("alexis")
	if err != nil {
		log.Fatal("Failed to create FirstName:", err)
	}

	// Create LastName
	lastName, err := domain.NewLastName("nahan")
	if err != nil {
		log.Fatal("Failed to create LastName:", err)
	}

	// Create SocialProfile
	socialProfile, err := domain.NewSocialProfile(
		domain.SocialMediaURL("instagram"),
		"https://instagram.com/alnah",
	)
	if err != nil {
		log.Fatal("Failed to create SocialProfile:", err)
	}

	user, err := domain.NewUser(domain.NewUserParams{
		UserID:         userID,
		Username:       username,
		Email:          email,
		Roles:          []domain.Role{domain.RoleAdmin},
		FirstName:      firstName,
		LastName:       lastName,
		SocialProfiles: []domain.SocialProfile{socialProfile},
		Clock:          clock,
	})
	if err != nil {
		log.Fatal("Failed to create User:", err)
	}

	log.Println(user)
}
