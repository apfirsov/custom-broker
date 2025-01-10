package broker

type logger interface {
	Info(args ...any)
	Error(args ...any)
}
