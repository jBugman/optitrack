// NatNet packet parsing attempt (not finished)
package natnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"math"
)

type rawFrame struct {
	frameType   uint64
	frameNumber uint64
	markerSets  []markerSet
	unidMarkers []Vector3
	rigidBodies []RigidBody
	size        int
	bodyNames   map[string]int
}

type markerSet struct {
	Name    string
	Markers []Vector3
}

func (f rawFrame) RigidBodies() map[string]RigidBody {
	result := make(map[string]RigidBody, len(f.bodyNames))
	for name, id := range f.bodyNames {
		body := f.rigidBodies[id]
		body.Name = name
		result[name] = body
	}
	return result
}

func Parse(buf []byte) (f Frame, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Panicked during parsing:", r)
			log.Printf("%#v\n", buf)
			err = r.(error)
		}
	}()
	frame, err := parsePacket(buf)
	return frame, err
}

func parsePacket(buf []byte) (rawFrame, error) {
	var packet rawFrame
	var offset int

	packet.frameType, _ = binary.Uvarint(buf[offset : offset+2])
	offset += 2
	if packet.frameType != 7 { // 7 - mocap data
		return packet, errors.New("Not mocap data packet")
	}
	packet.bodyNames = make(map[string]int)

	//	numBytes, _ := binary.Uvarint(buf[offset : offset+2]) // unknown nature
	offset += 2
	packet.frameNumber, _ = binary.Uvarint(buf[offset : offset+4])
	offset += 4
	markerSetCount, _ := binary.Uvarint(buf[offset : offset+4])
	offset += 4

	packet.rigidBodies = make([]RigidBody, markerSetCount)

	// markersets
	var markerCount uint64
	for ms := 0; ms < int(markerSetCount); ms++ {
		// Reading c-string
		bb := new(bytes.Buffer)
		var i int
		for i = 0; buf[offset+i] != 0; i++ {
			bb.WriteByte(buf[offset+i])
		}
		offset += i + 1
		name := bb.String()
		mSet := markerSet{Name: name}
		if name != "all" {
			packet.bodyNames[name] = ms
		}

		markerCount, _ = binary.Uvarint(buf[offset : offset+4])
		offset += 4
		for i = 0; i < int(markerCount); i++ {
			x := FloatFromBytes(buf[offset : offset+4])
			offset += 4
			y := FloatFromBytes(buf[offset : offset+4])
			offset += 4
			z := FloatFromBytes(buf[offset : offset+4])
			offset += 4
			v := Vector3{x, y, z}
			mSet.Markers = append(mSet.Markers, v)
		}
		packet.markerSets = append(packet.markerSets, mSet)
	}

	// unidentified markers
	unidMarkerCount, _ := binary.Uvarint(buf[offset : offset+4])
	offset += 4
	for i := 0; i < int(unidMarkerCount); i++ {
		x := FloatFromBytes(buf[offset : offset+4])
		offset += 4
		y := FloatFromBytes(buf[offset : offset+4])
		offset += 4
		z := FloatFromBytes(buf[offset : offset+4])
		offset += 4
		v := Vector3{x, y, z}
		packet.unidMarkers = append(packet.unidMarkers, v)
	}

	// rigid bodies
	rigidBodyCount, _ := binary.Uvarint(buf[offset : offset+4])
	offset += 4
	for i := 0; i < int(rigidBodyCount); i++ {
		id, _ := binary.Uvarint(buf[offset : offset+4])
		offset += 4
		x := FloatFromBytes(buf[offset : offset+4])
		offset += 4
		y := FloatFromBytes(buf[offset : offset+4])
		offset += 4
		z := FloatFromBytes(buf[offset : offset+4])
		offset += 4
		qx := FloatFromBytes(buf[offset : offset+4])
		offset += 4
		qy := FloatFromBytes(buf[offset : offset+4])
		offset += 4
		qz := FloatFromBytes(buf[offset : offset+4])
		offset += 4
		qw := FloatFromBytes(buf[offset : offset+4])
		offset += 4
		// associated marker positions
		nRigidMarkers, _ := binary.Uvarint(buf[offset : offset+4])
		offset += 4
		offset += int(nRigidMarkers) * 3 * 4
		// associated marker IDs
		offset += int(nRigidMarkers) * 4
		// associated marker sizes
		offset += int(nRigidMarkers) * 4
		// Mean marker error
		offset += 4
		// params
		offset += 2

		body := RigidBody{ID: int(id), Position: Vector3{x, y, z}, Rotation: Quaternion{qx, qy, qz, qw}}
		packet.rigidBodies[body.ID-1] = body
	}

	packet.size = -1 // TODO
	return packet, nil
}

func FloatFromBytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}
