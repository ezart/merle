package merle

type Storker interface {
	NewThinger(model string, demo bool) (Thinger, error)
}
