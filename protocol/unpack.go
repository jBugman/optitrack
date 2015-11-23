package natnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

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
