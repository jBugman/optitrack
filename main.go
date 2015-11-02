package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"time"

	"golang.org/x/net/ipv4"
)

const (
	multicastAddress = "239.255.42.99:1511"
	bufferSize       = 5000
)

func address() *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp4", multicastAddress)
	if err != nil {
		log.Fatal(err)
	}
	return addr
}

func record() error {
	log.Println("Recording packets")
	conn, err := net.ListenMulticastUDP("udp4", nil, address())
	if err != nil {
		return err
	}
	defer conn.Close()
	filename := time.Now().Format("2005-01-02_15-04.bin")
	outfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outfile.Close()
	defer outfile.Sync()

	recieved := 0
	buf := make([]byte, bufferSize)
	for i := 0; true; i++ {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}
		outfile.Write(buf)
		recieved += n
		if i%100 == 0 {
			log.Printf("Recieved %d bytes\n", recieved)
		}
	}
	return nil
}

type frameInfo struct {
	frameType     uint64
	frameNumber   uint64
	datasetsCount uint64
	size          int
}

// Packet parsing attempt (not finished)
func parsePacket(buf []byte) (frameInfo, error) {
	var packet frameInfo
	var offset int

	packet.frameType, _ = binary.Uvarint(buf[offset : offset+2])
	offset += 2
	if packet.frameType != 7 { // 7 - mocap data
		return packet, errors.New("Not mocap data packet")
	}

	//	numBytes, _ := binary.Uvarint(buf[offset : offset+2]) // unknown nature
	offset += 2
	packet.frameNumber, _ = binary.Uvarint(buf[offset : offset+4])
	offset += 4
	packet.datasetsCount, _ = binary.Uvarint(buf[offset : offset+4])
	offset += 4
	fmt.Println(packet)

	var numMarkers int
	var markers []byte
	for ms := 0; ms < int(packet.datasetsCount); ms++ {
		bb := new(bytes.Buffer)
		var i int
		for i = 0; buf[offset+i] != 0; i++ {
			bb.WriteByte(buf[offset+i])
		}
		offset += i + 1
		fmt.Println(bb.String())

		nm, _ := binary.Uvarint(buf[offset : offset+4])
		numMarkers = int(nm)
		offset += 4
		nbytes := numMarkers * 3 * 4
		markers = buf[offset : offset+nbytes]
		offset += nbytes
		fmt.Println(numMarkers, nbytes, markers, buf[offset:])
		for i = 0; i < numMarkers; i++ {
			var x, y, z float64
			off := 0
			x0, _ := binary.Uvarint(markers[off : off+4])
			x = math.Float64frombits(x0)
			off += 4
			y0, _ := binary.Uvarint(markers[off : off+4])
			y = math.Float64frombits(y0)
			off += 4
			z0, _ := binary.Uvarint(markers[off : off+4])
			z = math.Float64frombits(z0)
			fmt.Println(x, y, z)
		}
	}
	packet.size = offset
	return packet, nil
}

func play(filename string, fps int) error {
	log.Printf("Replaying %s in %d fps", filename, fps)

	datafile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer datafile.Close()

	addr := address()
	c, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		return err
	}
	defer c.Close()
	// Some low-lvl syscall magic to allow multicast on localhost
	conn := ipv4.NewPacketConn(c)
	conn.SetMulticastLoopback(true)
	defer conn.Close() // not sure if should close wrapper, no errors both ways

	frameTime := time.Duration(1000000.0/float64(fps)) * time.Microsecond
	log.Println("Using frame time", frameTime)

	buf := make([]byte, bufferSize)
	var frameCount int
	for frameCount = 0; true; frameCount++ {
		_, err := datafile.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}
		//		packet, err := parsePacket(buf)
		//		fmt.Println(buf[:packet.size])
		conn.WriteTo(buf, nil, addr)
		time.Sleep(frameTime) // Replaying in given fps
	}
	log.Printf("Played %d frames\n", frameCount)

	return nil
}

func main() {
	var recordMode = flag.Bool("record", false, "Run in record mode")
	var replayMode = flag.Bool("replay", false, "Replay recorded data")
	var fps = flag.Int("fps", 30, "Replay FPS")
	flag.Parse()

	switch {
	case *recordMode:
		record()
	case *replayMode:
		if len(flag.Args()) == 0 {
			log.Println("Filename required")
		} else {
			play(flag.Arg(0), *fps)
		}
	default:
		log.Println("Select mode (-help)")
	}
}
