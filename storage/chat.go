package storage

import (
	"encoding/json"
	"os"
	"sync"

	"wifi-chat/models"
)

const chatFile = "data/chats.json"

var mutex = &sync.RWMutex{}

func init() {
	// Buat folder data jika belum ada
	os.MkdirAll("data", 0755)
}

func SaveMessage(msg models.Message) error {
	mutex.Lock()
	defer mutex.Unlock()

	messages, _ := loadMessages()
	messages = append(messages, msg)

	// Batasi maksimal 500 pesan terakhir
	if len(messages) > 500 {
		messages = messages[len(messages)-500:]
	}

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(chatFile, data, 0644)
}

func LoadMessages() ([]models.Message, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	return loadMessages()
}

func loadMessages() ([]models.Message, error) {
	var messages []models.Message

	data, err := os.ReadFile(chatFile)
	if err != nil {
		if os.IsNotExist(err) {
			return messages, nil
		}
		return nil, err
	}

	err = json.Unmarshal(data, &messages)
	return messages, err
}

func ClearMessages() error {
	mutex.Lock()
	defer mutex.Unlock()
	return os.WriteFile(chatFile, []byte("[]"), 0644)
}
