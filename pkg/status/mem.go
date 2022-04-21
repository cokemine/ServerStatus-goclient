//go:build !linux
// +build !linux

package status

import "github.com/shirou/gopsutil/v3/mem"

func Memory() (uint64, uint64, uint64, uint64) {
	memory, _ := mem.VirtualMemory()
	swap, _ := mem.SwapMemory()
	return memory.Total / 1024.0, memory.Used / 1024.0, swap.Total / 1024.0, swap.Used / 1024.0
}
