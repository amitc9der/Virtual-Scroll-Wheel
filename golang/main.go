package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// For development, you might want to check the origin and allow it if it's your local environment
		// Here, we're allowing all origins for simplicity, which isn't recommended for production
		return true
	},
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Virtual Mouse</title>
    <script src="https://unpkg.com/htmx.org@1.9.6"></script>
    <style>
        #disk {
            width: 200px;
            height: 200px;
            border-radius: 50%;
            background-color: gray;
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            cursor: grab;
        }
    </style>
    <script>
		function calculateRotation() {
			// Simplified for example, replace with actual calculation logic
			let startAngle = 0; // You'd need to manage this state somehow
			let currentAngle = 0; // Ditto
			// ... calculation logic here ...
			return (currentAngle - startAngle) * (180 / Math.PI);
		}
    </script>
</head>
<body>
	<div id="disk" hx-post="/rotate" hx-trigger="mousemove from:#disk" hx-swap="none" hx-vals='{"rotation": "{{calculateRotation()}}"}'></div>
</body>
</html>
`))

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTemplate.Execute(w, r.Host)
}

func echo(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Connection closed: %v", err)
			} else {
				log.Printf("Read error: %v", err)
			}
			return
		}
		log.Printf("Received: %s", message)
		// Here you would handle the mouse scroll based on the rotation data
		// For now, we're just echoing back the message
		err = conn.WriteMessage(mt, message)
		if err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}

func rotateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body := make([]byte, r.ContentLength)
	r.Body.Read(body)
	rotation := strings.TrimSpace(string(body))

	// Convert rotation to float for further processing
	if rotationFloat, err := strconv.ParseFloat(rotation, 64); err == nil {
		log.Printf("Received rotation: %f", rotationFloat)
		// Here, you would handle the float value for mouse scrolling or other actions
	} else {
		log.Printf("Error parsing rotation: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", echo)
	http.HandleFunc("/rotate", rotateHandler)

	fmt.Println("Server is running on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
