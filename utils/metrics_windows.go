//go:build windows

package utils

import (
	"syscall"
	"unsafe"
)

var (
	psapi                = syscall.NewLazyDLL("psapi.dll")
	getProcessMemoryInfo = psapi.NewProc("GetProcessMemoryInfo")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	getCurrentProcess    = kernel32.NewProc("GetCurrentProcess")
)

type PROCESS_MEMORY_COUNTERS struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

func GetRSSBytes() (uint64, error) {
	handle, _, _ := getCurrentProcess.Call()
	var counters PROCESS_MEMORY_COUNTERS
	counters.CB = uint32(unsafe.Sizeof(counters))
	ret, _, err := getProcessMemoryInfo.Call(handle, uintptr(unsafe.Pointer(&counters)), uintptr(counters.CB))
	if ret == 0 {
		return 0, err
	}
	return uint64(counters.WorkingSetSize), nil
}

func GetCPUSec() (float64, error) {
	handle, _, _ := getCurrentProcess.Call()
	var creationTime, exitTime, kernelTime, userTime syscall.Filetime
	ret, _, err := kernel32.NewProc("GetProcessTimes").Call(
		handle,
		uintptr(unsafe.Pointer(&creationTime)),
		uintptr(unsafe.Pointer(&exitTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)
	if ret == 0 {
		return 0, err
	}
	kTime := uint64(kernelTime.HighDateTime)<<32 | uint64(kernelTime.LowDateTime)
	uTime := uint64(userTime.HighDateTime)<<32 | uint64(userTime.LowDateTime)
	totalSeconds := float64(kTime+uTime) / 10000000.0
	return totalSeconds, nil
}
