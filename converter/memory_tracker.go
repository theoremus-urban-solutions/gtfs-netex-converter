package converter

import (
	"runtime"
	"sync"
)

// MemoryTracker tracks memory usage during conversion
type MemoryTracker struct {
	mu              sync.Mutex
	peakMemoryMB    float64
	currentMemoryMB float64
	enabled         bool
}

// NewMemoryTracker creates a new memory tracker
func NewMemoryTracker(enabled bool) *MemoryTracker {
	return &MemoryTracker{
		enabled: enabled,
	}
}

// Sample records current memory usage and updates peak
func (mt *MemoryTracker) Sample() {
	if !mt.enabled {
		return
	}

	mt.mu.Lock()
	defer mt.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Convert bytes to megabytes
	mt.currentMemoryMB = float64(m.Alloc) / 1024 / 1024

	if mt.currentMemoryMB > mt.peakMemoryMB {
		mt.peakMemoryMB = mt.currentMemoryMB
	}
}

// GetCurrent returns current memory usage in MB
func (mt *MemoryTracker) GetCurrent() float64 {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	return mt.currentMemoryMB
}

// GetPeak returns peak memory usage in MB
func (mt *MemoryTracker) GetPeak() float64 {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	return mt.peakMemoryMB
}

// ForceGC forces garbage collection and samples memory
func (mt *MemoryTracker) ForceGC() {
	if !mt.enabled {
		return
	}
	runtime.GC()
	mt.Sample()
}
