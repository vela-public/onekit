package filekit

type Logger interface {
	Errorf(string, ...any)
	Warnf(string, ...any)
	Debugf(string, ...any)
	Infof(string, ...any)
}
