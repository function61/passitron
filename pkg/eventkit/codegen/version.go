package codegen

// struct to use for rendering version template
type versionData struct {
	Version string
}

func NewVersionData(version string) *versionData {
	return &versionData{version}
}
