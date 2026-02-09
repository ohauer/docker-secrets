// +build windows

package filewriter

// MaxPathLen is the maximum path length for Windows
// Using legacy MAX_PATH (260) for compatibility
// Extended paths (\\?\) and UNC paths are not supported for security
const MaxPathLen = 260
