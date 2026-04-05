package invalid

// Store — интерфейс хранилища невалидных метрик.
type Store interface {
	Save(rawPayload interface{}, errorReason string) error
}
