package thermal

import (
	"github.com/Erexo/Ventana/core/dto"
	"github.com/Erexo/Ventana/core/entity"
	"github.com/pkg/errors"
)

type ThermalBlock struct {
	arr  []dto.Point
	next int
	full bool
}

func CreateThermalBlock(size int) *ThermalBlock {
	return &ThermalBlock{
		arr:  make([]dto.Point, size),
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
	t.arr[t.next] = dto.CreatePoint(celsius, timestamp)
	t.next++
	if t.next >= len(t.arr) {
		t.next = 0
		t.full = true
	}
}

func (t *ThermalBlock) Last() (dto.Point, error) {
	if t.next == 0 {
		if t.full {
			return t.arr[len(t.arr)-1], nil
		}
		return dto.Point{}, errors.New("ThermalBlock is empty")
	}
	return t.arr[t.next-1], nil
}

func (t *ThermalBlock) Read() []dto.Point {
	if !t.full {
		ret := make([]dto.Point, t.next)
		copy(ret, t.arr[:t.next])
		return ret
	}
	ret := make([]dto.Point, len(t.arr))
	if t.next == 0 {
		copy(ret, t.arr)
	} else {
		copy(ret, t.arr[t.next:])
		copy(ret[len(t.arr)-t.next:], t.arr[:t.next])
	}
	return ret
}
