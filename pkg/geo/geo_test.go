package geo

import (
	"fmt"
	"testing"
)

var distanceTests = []struct {
	in  [2]Coord
	out float64
}{
	{
		in:  [2]Coord{{43.482928, -80.535819}, {43.482928, -80.535819}},
		out: 0,
	},
	{
		in:  [2]Coord{{43.481940, -80.537687}, {43.481979, -80.537590}},
		out: 10.821490633803409,
	},
	{
		in:  [2]Coord{{43.481979, -80.537590}, {43.481940, -80.537687}},
		out: 10.821490633803409,
	},
	{
		in:  [2]Coord{{43.482928, -80.535819}, {43.482889, -80.535771}},
		out: 5.390780229590655,
	},
}

func TestDistance(t *testing.T) {
	for _, tt := range distanceTests {
		t.Run(fmt.Sprintf("%v, %v = %v", tt.in[0], tt.in[1], tt.out), func(t *testing.T) {
			dist := Distance(tt.in[0], tt.in[1])

			if dist != tt.out {
				t.Errorf(`expected %v but got: %v`, tt.out, dist)
			}
		})
	}
}
