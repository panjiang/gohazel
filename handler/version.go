package handler

// ToSemver convert to semver
func ToSemver(version string) string {
	if len(version) > 0 && version[0] != 'v' {
		version = "v" + version
	}
	return version
}
