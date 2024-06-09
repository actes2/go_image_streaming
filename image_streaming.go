package image_streaming

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"reflect"
	"strconv"

	"fmt"
	"log"
	"net"
	"os"

	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/kbinani/screenshot"
)

const PACKET_SIZE = 65000

/*
Summary:

# This function is responsible for using our internal RUDP protcol to stream data

Parameters:

packet_Max -> int: Maximum amount of bits we'll be sending

connection -> *net.UDPConn: The UDP Connection object

img_Buffer -> *bytes.Buffer: The Buffer we'll be sending across our protcol
*/
func Transmit_UDP(packet_Max int, connection *net.UDPConn, img_Buffer *bytes.Buffer) bool {
	if send_Header(packet_Max, connection) {
		send_Payload(packet_Max, connection, img_Buffer)
	} else {
		log.Println("Header failed to send")
		return false
	}
	return true
}

// Sends a 'body' payload of bits to our client, until we empty the buffer
func send_Payload(packet_Max int, conn *net.UDPConn, b_image *bytes.Buffer) {
	response_buff := make([]byte, 16)

	for i, loader := 0, 0; i < packet_Max; i++ {

		if i == packet_Max-1 {
			conn.Write(b_image.Bytes()[loader:])
		} else {
			conn.Write(b_image.Bytes()[loader : loader+PACKET_SIZE])
		}

		loader += PACKET_SIZE

		conn.Read(response_buff)
	}

	b_image.Reset()
}

// Sends a 'header' packet that introduces ourselves and denotes our transfer max for the follow-up payload
func send_Header(packet_Max int, conn *net.UDPConn) bool {
	//println("\nPacket max:", packet_Max)

	header := fmt.Sprintf(`Z!mZ00mP:%d:`, packet_Max)

	//log.Println(header)
	conn.Write([]byte(header))

	response_buff := make([]byte, 16)

	conn.SetReadDeadline(time.Now().Add(400 * time.Millisecond))

	n, err := conn.Read(response_buff)
	if err != nil {
		log.Println("It would appear we timed out: ", err)
		return false
	}
	if n > 0 {
		if strings.Contains(string(response_buff), "ack") {
			//println("Ack")
			return true
		}
	}
	return false
}

// Listens for a header and payload with our image data.
func Listen_for_UDP_stream(
	port string,
	mute_io ...interface{},
) image.Image {

	var addr string

	if port == "" {
		addr = ":42069"
	} else {
		addr = ":" + port
	}
	// Define the server address and port

	udpAddr, err := net.ResolveUDPAddr("udp", addr)

	if err != nil {
		log.Println("Error resolving UDP address:", err)
		os.Exit(1)
	}

	// Create a UDP socket
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println("Error listening on UDP port:", err)
		os.Exit(1)
	}
	defer conn.Close()

	if len(mute_io) > 0 {
		log.Println("UDP server listening on", addr)
	}

	initial_buffer := make([]byte, 1024)

	// file, err := os.Create("received_image.jpg")
	// if err != nil {
	// 	log.Println("Error creating file:", err)
	// 	return
	// }
	// defer file.Close()

	for {
		// Read from UDP connection
		n, remoteAddr, err := conn.ReadFromUDP(initial_buffer)
		if err != nil {
			log.Println("Error reading from UDP:", err)
			continue
		}

		if n > 0 {
			// If we receive a packet and also the 'token' we anticipate.
			if strings.Contains(string(initial_buffer), "Z!mZ00m") {
				unpack := strings.Split(string(initial_buffer), ":")

				packet_Max, _ := strconv.Atoi(unpack[1])

				conn.WriteToUDP([]byte("ack"), remoteAddr)

				conn.SetReadDeadline(time.Now().Add(1 * time.Second))

				var stream_buff []byte

				for i := 0; i < packet_Max; i++ {
					packet_Buffer := make([]byte, PACKET_SIZE)

					conn.Read(packet_Buffer)

					stream_buff = append(stream_buff, packet_Buffer...)

					conn.WriteToUDP([]byte("ack"), remoteAddr)
				}

				imgreader := bytes.NewReader(bytes.TrimRight(stream_buff, "\x00\x00"))

				img := func() image.Image { read_img, _, _ := image.Decode(imgreader); return read_img }
				return img()
				//file.Write(stream_buff)
			}

		}

	}
}

/*
Screenshot and encode as a jpeg to a buffer.

Parameters:

Buffer -> *bytes.buffer - this is the image buffer.

capture_Region -> interface{} - this is a struct for a rect

quality -> int - this is our quality modifier for the jpeg.

resizer -> ...int - this is an optional parameter that allows for us to scale the image by a factor. 4 is basically real-time with a grosser quality.

Output:

This entire function returns true if it succeededs and false if it failes.
*/
func Screenshot_and_encode_jpeg_compressed(
	buffer *bytes.Buffer,
	capture_Region interface{},
	quality int,
	resizer int,
) bool {
	// Fortunately 1920x1080 in packets is small enough for us to shotgun the buffer.
	// Meaning this is basically live-time (especially with the resize)

	reg := reflect.ValueOf(capture_Region) // Reflect our interface

	img, _ := screenshot.Capture(int(reg.FieldByName("x").Int()), int(reg.FieldByName("y").Int()), int(reg.FieldByName("w").Int()), int(reg.FieldByName("h").Int()))

	if resizer <= 0 {
		resizer = 2
	}

	processed_img := imaging.Resize(img, img.Bounds().Dx()/resizer, img.Bounds().Dy()/resizer, imaging.Lanczos)

	err := jpeg.Encode(buffer, processed_img, &jpeg.Options{Quality: quality})
	if err != nil {
		print("Error occurred:", err)
		return false
	}

	return true
}

/*
Screenshot and encode as a jpeg to a buffer.

Parameters:

Buffer -> *bytes.buffer - this is the image buffer.

capture_Region -> interface{} - this is a struct for a rect

resizer -> ...int - this is an optional parameter that allows for us to scale the image by a factor. 4 is basically real-time with a grosser quality.

Output:

This entire function returns true if it succeededs and false if it failes.
*/
func Screenshot_and_encode_png_compressed(
	buffer *bytes.Buffer,
	capture_Region interface{},
	resizer ...int,
) bool {
	// Fortunately 1920x1080 in packets is small enough for us to shotgun the buffer.
	// Meaning this is basically live-time (especially with the resize)

	reg := reflect.ValueOf(capture_Region) // Reflect our interface

	img, _ := screenshot.Capture(int(reg.FieldByName("x").Int()), int(reg.FieldByName("y").Int()), int(reg.FieldByName("w").Int()), int(reg.FieldByName("h").Int()))

	if len(resizer) != 0 {
		processed_img := imaging.Resize(img, img.Bounds().Dx()/resizer[0], img.Bounds().Dy()/resizer[0], imaging.Lanczos)

		err := png.Encode(buffer, processed_img)
		if err != nil {
			print("Error occurred:", err)
			return false
		}

	} else { // No resizing
		err := png.Encode(buffer, img)
		if err != nil {
			print("Error occurred:", err)
			return false
		}
	}

	return true
}

// This function captures the whole display.
func Screenshot_and_encode_png(buffer *bytes.Buffer) bool {
	img, _ := screenshot.CaptureDisplay(0)
	err := png.Encode(buffer, img)
	//err := jpeg.Encode(buff, img, &jpeg.Options{Quality: 5})
	if err != nil {
		print("Error occurred:", err)
		return false
	}
	return true
}

// This function captures the whole display.
func Screenshot_and_encode_jpeg(buffer *bytes.Buffer, quality ...int) bool {
	img, _ := screenshot.CaptureDisplay(0)

	var jpeg_quality int

	if len(quality) == 0 {
		jpeg_quality = 5
	} else {
		jpeg_quality = quality[0]
	}

	err := jpeg.Encode(buffer, img, &jpeg.Options{Quality: jpeg_quality})
	if err != nil {
		print("Error occurred:", err)
		return false
	}
	return true
}
