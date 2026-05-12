package persistence

type RDBState struct {
	BGSaveRunning bool
	DBCopy        map[string]string
}
