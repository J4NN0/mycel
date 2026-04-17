package loggertest

type Logger struct{}

func (m *Logger) Printf(_ string, _ ...interface{})   {}
func (m *Logger) Warningf(_ string, _ ...interface{}) {}
func (m *Logger) Fatalf(_ string, _ ...interface{})   {}
