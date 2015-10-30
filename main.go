package main

import (
	"log"
	"net"
	//	"fmt"
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp4", "239.255.42.99:1511")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	outfile, err := os.Create("outfile.bin")
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()
	defer outfile.Sync()

	buf := make([]byte, 5000)
	for {
		_, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
		}
		fmt.Println(0)
		//fmt.Println(buf)
		outfile.Write(buf)

		var offset int
		iMessage, _ := binary.Uvarint(buf[offset : offset+2])
		offset += 2
		//numBytes, _ := binary.Uvarint(buf[offset:offset+2])
		offset += 2
		if iMessage != 7 { // 7 - mocap data
			continue
		}
		//frame, _ := binary.Uvarint(buf[offset:offset+4])
		offset += 4
		numDataSets, _ := binary.Uvarint(buf[offset : offset+4])
		offset += 4
		//fmt.Println(iMessage, numBytes, frame, numDataSets)
		var numMarkers int
		var markers []byte
		for ms := 0; ms < int(numDataSets); ms++ {
			bb := new(bytes.Buffer)
			var i int
			for i = 0; buf[offset+i] != 0; i++ {
				bb.WriteByte(buf[offset+i])
			}
			offset += i + 1
			//fmt.Println(bb.String())
			nm, _ := binary.Uvarint(buf[offset : offset+4])
			numMarkers = int(nm)
			offset += 4
			nbytes := numMarkers * 3 * 4
			markers = buf[offset : offset+nbytes]
			offset += nbytes
			//fmt.Println(numMarkers, nbytes, markers, buf[offset:])
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
				//fmt.Println(x, y, z)
				_, _, _ = x, y, z
			}
		}
	}
}
