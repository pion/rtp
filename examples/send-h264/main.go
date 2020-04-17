// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

// send-h264 demonstrates how to send and receieve RTP packets over the network.
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
)

func readThread() int {
	conn, err := net.ListenUDP("udp", &(net.UDPAddr{
		Port: 0,
		IP:   net.ParseIP("0.0.0.0"),
	}))
	if err != nil {
		panic(err)
	}

	go func() {
		buf := make([]byte, 1024)
		var pkt rtp.Packet

		for {
			n, remoteAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				panic(err)
			}

			if err = pkt.Unmarshal(buf[:n]); err != nil {
				panic(err)
			}

			fmt.Printf("Received seq_num %d from %s\n", pkt.SequenceNumber, remoteAddr)
		}
	}()
	if udpAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		return udpAddr.Port
	}

	return 0
}

func main() {
	data, err := os.ReadFile("output.h264")
	if err != nil {
		log.Fatal(err)
	}

	listeningPort := readThread()
	conn, err := net.DialUDP("udp", nil, &(net.UDPAddr{
		Port: listeningPort,
		IP:   net.ParseIP("0.0.0.0"),
	}))
	if err != nil {
		panic(err)
	}

	packetizer := rtp.NewPacketizer(100, 98, 0x1234ABCD, &codecs.H264Payloader{}, rtp.NewRandomSequencer(), 90000)

	// A real application would call Packetize in a loop with proper timing.
	// For this demo we aren't actually timing things
	packets := packetizer.Packetize(data, 0)

	for i := range packets {
		marshaled, err := packets[i].Marshal()
		if err != nil {
			panic(err)
		}
		if _, err = conn.Write(marshaled); err != nil {
			panic(err)
		}
		time.Sleep(time.Millisecond * 100)
	}
}
