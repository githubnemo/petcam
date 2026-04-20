package main

import (
    "fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"
    "flag"
)

var (
	clients   = make(map[chan []byte]bool)
	clientsMu sync.Mutex
    flag_audioDevice = flag.String("audioDevice", "plughw:0,0", "The ALSA device identifier, e.g. plughw:0,0.")
    flag_port = flag.Int("port", 8000, "The TCP port to serve the HTTP stream on.")
)

func addClient(ch chan []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	clients[ch] = true
}

func removeClient(ch chan []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	delete(clients, ch)
	close(ch)
}

func broadcaster() {
	cmd := exec.Command("ffmpeg",
		"-f", "alsa",
		"-i", *flag_audioDevice,
		"-ac", "1",
		"-ar", "16000",
		"-f", "mp3",
		"-b:a", "32k",
		"-fflags", "nobuffer",
		"-")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 2048)

	for {
		n, err := stdout.Read(buf)
		if err != nil {
			log.Println("ffmpeg stopped:", err)
			return
		}

		chunk := make([]byte, n)
		copy(chunk, buf[:n])

		clientsMu.Lock()
		for ch := range clients {
			select {
			case ch <- chunk:
			default:
				// drop slow clients
			}
		}
		clientsMu.Unlock()
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Transfer-Encoding", "chunked")
    w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ch := make(chan []byte, 10)
	addClient(ch)
	defer removeClient(ch)

	for data := range ch {
		_, err := w.Write(data)
		if err != nil {
			return
		}
		flusher.Flush()
	}
}

func main() {
    flag.Parse()

	go broadcaster()

	http.HandleFunc("/stream", handler)

	log.Println("Listening on", *flag_port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *flag_port), nil))
}
