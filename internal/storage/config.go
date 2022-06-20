package storage

type SaveConfig struct {
	ToMem      bool
	ToFile     bool
	ToFilePath string
	ToFileSync bool
	ToDatabase bool
}
