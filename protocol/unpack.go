// NatNet packet parsing attempt (not finished)
package natnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
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
}

type markerSet struct {
	Name    string
	Markers []Vector3
}

type Vector3 struct {
	X, Y, Z float32
}

func (v Vector3) String() string {
	return fmt.Sprintf("(%.5f, %.5f, %.5f)", v.X, v.Y, v.Z)
}

type Quaternion struct {
	X, Y, Z, W float32
}

func (v Quaternion) String() string {
	return fmt.Sprintf("(%.5f, %.5f, %.5f, %.5f)", v.X, v.Y, v.Z, v.W)
}

type Frame interface {
	RigidBodies() map[string]RigidBody
}

func (f rawFrame) RigidBodies() map[string]RigidBody {
	result := make(map[string]RigidBody, len(f.rigidBodies))
	for i := 0; i < len(f.rigidBodies); i++ {
		name := f.markerSets[i].Name
		f.rigidBodies[i].Name = name
		result[name] = f.rigidBodies[i]
	}
	return result
}

type RigidBody struct {
	ID       int
	Name     string
	Position Vector3
	Rotation Quaternion
}

func Parse(buf []byte) (f Frame, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Panicked during parsing:", r)
			log.Println(buf)
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

	//	numBytes, _ := binary.Uvarint(buf[offset : offset+2]) // unknown nature
	offset += 2
	packet.frameNumber, _ = binary.Uvarint(buf[offset : offset+4])
	offset += 4
	markerSetCount, _ := binary.Uvarint(buf[offset : offset+4])
	offset += 4

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
		mSet := markerSet{Name: bb.String()}

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
		packet.rigidBodies = append(packet.rigidBodies, body)
	}

	packet.size = -1 // TODO
	return packet, nil
}

func FloatFromBytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}
