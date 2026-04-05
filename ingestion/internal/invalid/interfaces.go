package invalid

type Store interface {
	Save(rawPayload interface{}, errorReason string) error
}
