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