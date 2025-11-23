package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"wifi-chat/handlers"
	"wifi-chat/storage"
)

func main() {
	hub := handlers.NewHub()
	go hub.Run()

	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Pages
	http.HandleFunc("/", handlers.ServeHome)

	// Auth API
	http.HandleFunc("/api/register", handlers.HandleRegister)
	http.HandleFunc("/api/login", handlers.HandleLogin)
	http.HandleFunc("/api/logout", handlers.HandleLogout)
	http.HandleFunc("/api/verify", handlers.HandleVerify)

	// Chat API
	http.HandleFunc("/api/clear", handleClearChat)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleWebSocket(hub, w, r)
	})

	ip := getLocalIP()
	port := "8080"

	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║              🚀 WiFi Chat Server Started                  ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Local:   http://localhost:%s                          ║\n", port)
	fmt.Printf("║  Network: http://%-40s ║\n", ip+":"+port)
	fmt.Println("╠═══════════════════════════════════════════════════════════╣")
	fmt.Println("║  📁 Data disimpan di folder: data/                        ║")
	fmt.Println("║     • users.json  - data akun                             ║")
	fmt.Println("║     • chats.json  - history chat                          ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

func handleClearChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := storage.ClearMessages()
	if err != nil {
		http.Error(w, "Failed to clear", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Chat cleared"))
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return getLocalIPFallback()
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func getLocalIPFallback() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}
