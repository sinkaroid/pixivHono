//go:build !windows

package utils

import (
	"syscall"
)

func GetRSSBytes() (uint64, error) {
	var rusage syscall.Rusage
	err := syscall.Getrusage(syscall.RUSAGE_SELF, &rusage)
	if err != nil {
		return 0, err
	}
	return uint64(rusage.Maxrss) * 1024, nil
}

func GetCPUSec() (float64, error) {
	var rusage syscall.Rusage
	err := syscall.Getrusage(syscall.RUSAGE_SELF, &rusage)
	if err != nil {
		return 0, err
	}
	userSec := float64(rusage.Utime.Sec) + float64(rusage.Utime.Usec)/1000000.0
	sysSec := float64(rusage.Stime.Sec) + float64(rusage.Stime.Usec)/1000000.0
	return userSec + sysSec, nil
}
