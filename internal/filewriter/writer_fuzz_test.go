package filewriter

import (
	"os"
	"testing"
)

func FuzzValidatePath(f *testing.F) {
	// Seed with valid paths
	f.Add("/tmp/test")
	f.Add("/var/lib/secrets/test.txt")
	f.Add("/home/user/secrets/key.pem")

	// Seed with attack vectors
	f.Add("../../../etc/passwd")
	f.Add("/tmp/../../../etc/passwd")
	f.Add("/tmp/./test")
	f.Add("relative/path")
	f.Add("/tmp/test\x00hidden")
	f.Add("/tmp/test\n/etc/passwd")
	f.Add("/tmp/test\r\n/etc/passwd")
	f.Add("//tmp//test")
	f.Add("/tmp/test/..")
	f.Add("/tmp/test/../..")
	f.Add("C:\\Windows\\System32")
	f.Add("/tmp/\u202e/test")

	f.Fuzz(func(t *testing.T, path string) {
		// Should not panic
		_ = validatePath(path)
	})
}

func FuzzValidateMode(f *testing.F) {
	// Seed with valid modes
	f.Add(uint32(0600))
	f.Add(uint32(0644))
	f.Add(uint32(0400))

	// Seed with invalid modes
	f.Add(uint32(0777))
	f.Add(uint32(0666))
	f.Add(uint32(0002))
	f.Add(uint32(0xFFFFFFFF))

	f.Fuzz(func(t *testing.T, mode uint32) {
		// Should not panic
		_ = validateMode(os.FileMode(mode))
	})
}
