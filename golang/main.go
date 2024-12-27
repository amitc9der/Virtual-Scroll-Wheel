package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"syscall"

	"github.com/gorilla/websocket"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procMouseEvent    = user32.NewProc("mouse_event")
	MOUSEEVENTF_WHEEL = 0x0800
)

type RotationData struct {
	Rotation float64 `json:"rotation"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Simulate mouse scrolling using the Windows API
func simulateScroll(scrollAmount int) {
	// Multiply by 120 to simulate one scroll tick
	scrollDelta := uintptr(scrollAmount * 120)
	procMouseEvent.Call(uintptr(MOUSEEVENTF_WHEEL), 0, 0, scrollDelta, 0)
}

// Serve the main HTML page
func servePage(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>HTMX Scroll Control</title>
  <script src="https://unpkg.com/htmx.org"></script>
  <style>
    body {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 100vh;
      background: #282c34;
      margin: 0;
      color: white;
      font-family: Arial, sans-serif;
    }
    #disk {
      width: 200px;
      height: 200px;
      background: radial-gradient(circle, #4b4b4b, #1f1f1f);
      border-radius: 50%;
      border: 4px solid #ccc;
      position: relative;
    }
  </style>
</head>
<body>
  <div id="disk" hx-get="/rotate" hx-trigger="mousemove" hx-vals='{"eventType":"move"}' hx-swap="none"></div>
  <script>
    const disk = document.getElementById('disk');
    let lastAngle = 0;

    function calculateAngle(event) {
      const rect = disk.getBoundingClientRect();
      const centerX = rect.left + rect.width / 2;
      const centerY = rect.top + rect.height / 2;
      const deltaX = event.clientX - centerX;
      const deltaY = event.clientY - centerY;
      return Math.atan2(deltaY, deltaX) * (180 / Math.PI);
    }

    disk.addEventListener('mousemove', (event) => {
      if (event.buttons === 1) {
        const currentAngle = calculateAngle(event);
        const deltaAngle = currentAngle - lastAngle;
        lastAngle = currentAngle;

        // Send rotation data to server
        const socket = new WebSocket('ws://localhost:8080/ws');
        socket.onopen = () => {
          socket.send(JSON.stringify({ rotation: deltaAngle }));
        };
      }
    });
  </script>
</body>
</html>
`
	tmplFuncs := template.Must(template.New("main").Parse(tmpl))
	tmplFuncs.Execute(w, nil)
}

// Handle WebSocket connection
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket Client connected")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket Read error:", err)
			break
		}

		var data RotationData
		if err := json.Unmarshal(message, &data); err != nil {
			log.Println("JSON Unmarshal error:", err)
			continue
		}

		// Simulate mouse scrolling
		scrollAmount := int(data.Rotation)
		simulateScroll(scrollAmount)
	}
}

func main() {
	// Serve static content and WebSocket handler
	http.HandleFunc("/", servePage)
	http.HandleFunc("/ws", handleWebSocket)

	// Start server
	port := ":6610"
	fmt.Println("Server running on http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
