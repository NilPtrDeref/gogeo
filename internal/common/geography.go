package common

//go:generate msgp -tests=false

type Counties []County

type County struct {
	Id    string        `msg:"id"`
	Name  string        `msg:"name"`
	State string        `msg:"state"`
	Parts []Coordinates `msg:"coordinates"`
}

type Coordinates []float32
