package outbox

// Event represents a message stored in the outbox table.
type Event struct {
	ID      int64
	Topic   string
	Payload []byte
}
