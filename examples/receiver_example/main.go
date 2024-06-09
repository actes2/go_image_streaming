package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/actes2/go_image_streaming"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type sharedImage struct {
	img image.Image
	mut sync.RWMutex
}

// This function generates a test image.
func generateImage() image.Image {

	img := image.NewRGBA(image.Rect(0, 0, 300, 300))

	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{uint8(time.Now().UnixNano() % 255), 0, 0, 255}}, image.Point{}, draw.Src)

	for x := range 300 {
		for y := range 300 {
			if x%2 == 1 {
				img.Set(x, x, color.RGBA{R: 0, B: uint8(x), G: 255, A: 0})
			}
			if x%2 == 0 {
				img.Set(x, y, color.RGBA{R: 30, B: 255, G: 255, A: 0})
			}
		}
	}

	return img
}

func serverFiles(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join("app", r.RequestURI)

	log.Println(r.RequestURI)

	http.ServeFile(w, r, path)
}

func serveImage(w http.ResponseWriter, r *http.Request, s_img *sharedImage) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error occurred:", err)
	}

	defer conn.Close()

	for {

		var buf bytes.Buffer

		img := s_img.img

		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100})
		if err != nil {
			log.Println("It would appear we had an error encoding the jpeg (Possibly a malformed packet!)")
			s_img.img = generateImage()
			return
		}

		err = conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
		if err != nil {
			if strings.Contains(err.Error(), "An established connection was aborted by the software") {
				log.Println("Client disconnected")
				return
			}

			log.Println("Error occurred:", err)
			return
		}
		//time.Sleep(time.Millisecond * 200)
	}

}

func main() {

	s_img := sharedImage{
		img: generateImage(),
		mut: sync.RWMutex{},
	}

	log.Println("Starting web-service on: localhost:4469\nStarting UDP socket on: localhost:4470")

	go func() {
		for {
			blah := go_image_streaming.Listen_for_UDP_stream("4470")
			s_img.mut.Lock()

			s_img.img = blah

			s_img.mut.Unlock()
		}
	}()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveImage(w, r, &s_img)
	})

	http.HandleFunc("/", serverFiles)

	err := http.ListenAndServe(":4469", nil)
	if err != nil {
		log.Fatal(err)
	}
}
