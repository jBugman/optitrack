package natnet

import "fmt"

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

type RigidBody struct {
	ID       int
	Name     string
	Position Vector3
	Rotation Quaternion
}
