package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"wifi-chat/models"
	"wifi-chat/storage"
)

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendAuthResponse(w, false, "Data tidak valid", "")
		return
	}

	// Validasi
	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 3 {
		sendAuthResponse(w, false, "Username minimal 3 karakter", "")
		return
	}
	if len(req.Username) > 20 {
		sendAuthResponse(w, false, "Username maksimal 20 karakter", "")
		return
	}
	if len(req.Password) < 4 {
		sendAuthResponse(w, false, "Password minimal 4 karakter", "")
		return
	}

	err := storage.Register(req.Username, req.Password)
	if err != nil {
		sendAuthResponse(w, false, err.Error(), "")
		return
	}

	sendAuthResponse(w, true, "Registrasi berhasil! Silakan login.", "")
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendAuthResponse(w, false, "Data tidak valid", "")
		return
	}

	token, err := storage.Login(req.Username, req.Password)
	if err != nil {
		sendAuthResponse(w, false, err.Error(), "")
		return
	}

	sendAuthResponse(w, true, "Login berhasil!", token)
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("Authorization")
	if token != "" {
		storage.Logout(token)
	}

	sendAuthResponse(w, true, "Logout berhasil", "")
}

func HandleVerify(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		sendAuthResponse(w, false, "Token tidak ditemukan", "")
		return
	}

	username, err := storage.ValidateToken(token)
	if err != nil {
		sendAuthResponse(w, false, "Token tidak valid", "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"username": username,
	})
}

func sendAuthResponse(w http.ResponseWriter, success bool, message, token string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AuthResponse{
		Success: success,
		Message: message,
		Token:   token,
	})
}
