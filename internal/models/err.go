package models

type QueueFullErr struct{}

func (e *QueueFullErr) Error() string {
	return "очередь переполнена"
}

type QueueNotFound struct{}

func (e *QueueNotFound) Error() string {
	return "очередь не найдена"
}
