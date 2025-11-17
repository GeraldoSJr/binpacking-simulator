package bin

import (
	"allocation-simulator/item"
)

// The bin struct stores the itens in a slice
type Bin struct {
	ID                   int
	CPU_Capacity         int64
	Memory_Capacity      int64
	Current_CPU_Usage    int64
	Current_Memory_Usage int64
	Closed               bool
}

func NewBin(id int, cpucapacity int64, memorycapacity int64) *Bin {
	return &Bin{
		ID:                   id,
		CPU_Capacity:         cpucapacity,
		Memory_Capacity:      memorycapacity,
		Current_CPU_Usage:    0,
		Current_Memory_Usage: 0,
		Closed:               false,
	}
}

func (b *Bin) addItem(it item.Item) {
	b.Current_CPU_Usage += it.CPU
	b.Current_Memory_Usage += it.Memory
}

func (b *Bin) RemoveItem(it item.Item) {
	b.Current_CPU_Usage -= it.CPU
	b.Current_Memory_Usage -= it.Memory
}

func (b *Bin) RemainingCPUCapacity() int64 {
	return b.CPU_Capacity - b.Current_CPU_Usage
}

func (b *Bin) RemainingMemoryCapacity() int64 {
	return b.Memory_Capacity - b.Current_Memory_Usage
}

func (b *Bin) CanFit(it item.Item) bool {
	return (b.RemainingCPUCapacity() >= it.CPU) && (b.RemainingMemoryCapacity() >= it.Memory)
}

func (b *Bin) Empty() bool {
	return b.Current_CPU_Usage == 0 && b.Current_Memory_Usage == 0
}

func (b *Bin) TryAdd(it item.Item) bool {
    if b.Closed || !b.CanFit(it) {
        return false
    }

    b.addItem(it)
    return true
}
