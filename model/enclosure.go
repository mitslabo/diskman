package model

import (
	"fmt"

	"diskman/config"
)

type Enclosure struct {
	Def config.Enclosure
}

func NewEnclosure(def config.Enclosure) Enclosure {
	return Enclosure{Def: def}
}

func (e Enclosure) SlotAt(row, col int) (int, bool) {
	if row < 0 || row >= e.Def.Rows || col < 0 || col >= e.Def.Cols {
		return 0, false
	}
	return e.Def.Grid[row][col], true
}

func (e Enclosure) DevicePath(slot int) (string, bool) {
	k := fmt.Sprintf("%d", slot)
	v, ok := e.Def.Devices[k]
	return v, ok
}
