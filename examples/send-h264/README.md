# send-h264
send-h264 shows how to packetize H264 and send across the network via RTP

## Instructions
### Create H264 Annex-B file named `output.h264`
```
ffmpeg -f lavfi -i testsrc=duration=30:size=1280x720:rate=30 -c:v libx264 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -pix_fmt yuv420p output.h264
```

### Run send-h264
Make sure your `output.h264` is in the `send-h264` directory. Then run the example.


```
go run .
```

This results in output like
```
Received seq_num 31478 from 127.0.0.1:61169
Received seq_num 31479 from 127.0.0.1:61169
Received seq_num 31480 from 127.0.0.1:61169
Received seq_num 31481 from 127.0.0.1:61169
Received seq_num 31482 from 127.0.0.1:61169
Received seq_num 31483 from 127.0.0.1:61169
```

The receiver will print metadata about the RTP packets. We send a RTP packet every 100 milliseconds,
this is an arbitrary time. A real application would pace them correctly.
