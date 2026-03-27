//go:build !linux

package localsystem

// piModelInfo returns empty strings on non-Linux platforms.
func piModelInfo() (model, codename string) {
	return "", ""
}
