package limits

// DesiredLimits is a structure that specifies the limits desired by a running application
type DesiredLimits struct {
	FileLimitWant uint64
	FileLimitMin  uint64
}
