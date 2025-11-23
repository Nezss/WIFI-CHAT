package storage

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"wifi-chat/models"
)

const userFile = "data/users.json"

var (
	userMutex = &sync.RWMutex{}
	sessions  = make(map[string]string) // token -> username
	sessMutex = &sync.RWMutex{}
)

var (
	ErrUserExists   = errors.New("username sudah digunakan")
	ErrUserNotFound = errors.New("username tidak ditemukan")
	ErrWrongPass    = errors.New("password salah")
	ErrInvalidToken = errors.New("token tidak valid")
)

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func loadUsers() ([]models.User, error) {
	var users []models.User
	data, err := os.ReadFile(userFile)
	if err != nil {
		if os.IsNotExist(err) {
			return users, nil
		}
		return nil, err
	}
	err = json.Unmarshal(data, &users)
	return users, err
}

func saveUsers(users []models.User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(userFile, data, 0644)
}

func Register(username, password string) error {
	userMutex.Lock()
	defer userMutex.Unlock()

	users, err := loadUsers()
	if err != nil {
		return err
	}

	// Cek apakah username sudah ada
	for _, u := range users {
		if u.Username == username {
			return ErrUserExists
		}
	}

	// Buat user baru
	newUser := models.User{
		Username:  username,
		Password:  hashPassword(password),
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	users = append(users, newUser)
	return saveUsers(users)
}

func Login(username, password string) (string, error) {
	userMutex.RLock()
	defer userMutex.RUnlock()

	users, err := loadUsers()
	if err != nil {
		return "", err
	}

	for _, u := range users {
		if u.Username == username {
			if u.Password == hashPassword(password) {
				// Generate token
				token := generateToken()
				sessMutex.Lock()
				sessions[token] = username
				sessMutex.Unlock()
				return token, nil
			}
			return "", ErrWrongPass
		}
	}
	return "", ErrUserNotFound
}

func ValidateToken(token string) (string, error) {
	sessMutex.RLock()
	defer sessMutex.RUnlock()

	username, ok := sessions[token]
	if !ok {
		return "", ErrInvalidToken
	}
	return username, nil
}

func Logout(token string) {
	sessMutex.Lock()
	defer sessMutex.Unlock()
	delete(sessions, token)
}

func GetOnlineUsers() []string {
	sessMutex.RLock()
	defer sessMutex.RUnlock()

	users := make([]string, 0, len(sessions))
	seen := make(map[string]bool)
	for _, username := range sessions {
		if !seen[username] {
			users = append(users, username)
			seen[username] = true
		}
	}
	return users
}
