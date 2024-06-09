package main

import (
	"bytes"
	"fmt"
	"math"
	"net"

	"github.com/actes2/go_image_streaming"
)

func main() {
	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:4470")

	if err != nil {
		fmt.Println("We've failed to resolve the udp address")
		return
	}

	img_buff := new(bytes.Buffer)

	conn, _ := net.DialUDP("udp", nil, udpAddr)
	defer conn.Close()

	for {

		// This is gross as I just threw in an in-line struct, but you should be able to feed it any interface motif.
		go_image_streaming.Screenshot_and_encode_jpeg_compressed(img_buff, struct {
			x int
			y int
			w int
			h int
		}{
			x: 0,
			y: 0,
			w: 1000,
			h: 1000,
		},
			50,
			4)
		buff_Max := len(img_buff.Bytes())
		packet_Max := int(math.Ceil((float64(buff_Max) / float64(go_image_streaming.PACKET_SIZE))))

		println(buff_Max)

		go_image_streaming.Transmit_UDP(packet_Max, conn, img_buff)
		//time.Sleep(time.Second * 1)

	}
}
