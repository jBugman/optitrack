package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"time"
)

const multicastAddress = "239.255.42.99:1511"

func record() error {
	log.Println("Recording packets")
	addr, err := net.ResolveUDPAddr("udp4", multicastAddress)
	if err != nil {
		return err
	}
	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
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
	buf := make([]byte, 5000)
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

// Packet parsing attempt (not finished)
func parsePacket(buf []byte) {
	var offset int
	iMessage, _ := binary.Uvarint(buf[offset : offset+2])
	offset += 2
	numBytes, _ := binary.Uvarint(buf[offset : offset+2])
	offset += 2
	if iMessage != 7 { // 7 - mocap data
		return
	}
	frame, _ := binary.Uvarint(buf[offset : offset+4])
	offset += 4
	numDataSets, _ := binary.Uvarint(buf[offset : offset+4])
	offset += 4
	fmt.Println(iMessage, numBytes, frame, numDataSets)
	var numMarkers int
	var markers []byte
	for ms := 0; ms < int(numDataSets); ms++ {
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
}

func main() {
	record()
}
