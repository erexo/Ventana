package thermal

import "github.com/Erexo/Ventana/core/entity"

type Point struct {
	Celsius   entity.Temperature
	Timestamp entity.UnixTime
}

type ThermalBlock struct {
	arr  []Point
	next int
	full bool
}

func CreateThermalBlock(size int) *ThermalBlock {
	return &ThermalBlock{
		arr:  make([]Point, size),
		next: 0,
		full: false,
	}
}

func (t *ThermalBlock) IsFull() bool {
	return t.full
}

func (t *ThermalBlock) Len() int {
	if t.full {
		return len(t.arr)
	}
	return t.next
}

func (t *ThermalBlock) Add(celsius entity.Temperature, timestamp entity.UnixTime) {
	t.arr[t.next] = Point{celsius, timestamp}
	t.next++
	if t.next >= len(t.arr) {
		t.next = 0
		t.full = true
	}
}

func (t *ThermalBlock) Read() []Point {
	if !t.full {
		ret := make([]Point, t.next)
		copy(ret, t.arr[:t.next])
		return ret
	}
	ret := make([]Point, len(t.arr))
	if t.next == 0 {
		copy(ret, t.arr)
	} else {
		copy(ret, t.arr[t.next:])
		copy(ret[len(t.arr)-t.next:], t.arr[:t.next])
	}
	return ret
}
