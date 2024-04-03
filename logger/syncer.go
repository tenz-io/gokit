package logger

type NoopSyncer struct{}

func (ns *NoopSyncer) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (ns *NoopSyncer) Sync() error {
	return nil
}
