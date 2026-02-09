// +build linux darwin freebsd openbsd netbsd dragonfly solaris aix

package filewriter

// MaxPathLen is the maximum path length for Unix-like systems
const MaxPathLen = 4096 // PATH_MAX on most Unix systems
