# GO Image Streaming over UDP
This repository exists to make my personal UDP streaming library available to whoever wants to use it and also importable.

## Features

A coherent RUDP protocol at the most baseline level for transmitting image data across the internet.

The ability to specify size and dimensions of the screenshot

## Usage

I've found and tested that this module works well for web-based streaming at lower resolutions. My tests involved just making a socket between the 'server' and a javascript socket for dynamic refreshing.

In most tests anything less than 1200, 1200 pixels streams basically in real-time, while higher resolutions particularly in the 2000+ range end up chugging (I'm not sure of the principals behind fixing this, though I'd imagine big streaming studios and applications would settle for crunchy jpegs or highly optimized custom encoders)


To leverage this library just invoke the:
```go
Listen_for_UDP_stream(port string, mute_io ...interface{})
```
> port -> string: our port for listening to.

> mute_io -> ...interface{}: a conditonal flag for logging, leave it to 0 if you don't want logs.

function, which is used to await our special UDP packet from the sender.

Then transmit data to the Listener via a sender application of sorts - or the same application, do what you must I guess.

```go
func Screenshot_and_encode_jpeg_compressed(
	buffer *bytes.Buffer, // Our image buffer
	capture_Region interface{}, // the region interface which anticipates a rect-style object. x,y,w,h
	quality int, // quality in percents. I typically go with 6 for real-time but there's different permutations
	resizer int, // another quality reducer, I keep this at 2 as it divides the resolution and then blows it back up.
) bool
```

> buffer -> *bytes.buffer: This is the buffer we leverage for an image buffer.

> capture_Region -> interface{}: This anticipates an object with x,y,w,h parameters in its structure. I'm pretty sure this allows for plug-in-play rects but I could be incorrect, since my only actual tests are manually feeding structs to it.

> quality -> int: Takes in quality as a percent, with 100 being the best and 0 being the absolute worst.

> resizer -> int: Takes in a resizer factor, which I then under-the-hood resize back to its proper resolution, giving a 'fuzzier/blurrier' image but resulting in less packets which means better speed. I keep this at 2.

Additionally, there's a few other functions included that follow the same style as the one above for sending. Including a PNG counterpart to the example

## Examples

Sender Example: 

```go
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
        
        // The reasoning for this portion is because UDP is limited at 65,535 Bytes, forcing us to wrap around our buff
        buff_Max := len(img_buff.Bytes())
        packet_Max := int(math.Ceil((float64(buff_Max) / float64(go_image_streaming.PACKET_SIZE))))


        go_image_streaming.Transmit_UDP(packet_Max, conn, img_buff) // Sends off our packet header and body upon ack.


    }
}

```


Receiver Example:

```go
package main

import (
    "bytes"
    "image"
    "image/jpeg"
    "log"
    "os"
    "github.com/actes2/go_image_streaming"
)

func main() {
    
    // This hard exits, usually you'd wrap this with a goroutine
    img := go_image_streaming.Listen_for_UDP_stream("4470") 

    // Create a new file to save the PNG
    outFile, err := os.Create("output.jpeg")
    if err != nil {
        log.Fatalf("failed to create output file: %v", err)
    }
    defer outFile.Close()

    // Encode the image to PNG and write it to the file
    if err := jpeg.Encode(outFile, img); err != nil {
        log.Fatalf("failed to encode image to jpeg: %v", err)
    }

    log.Println("PNG image successfully created")
}
```