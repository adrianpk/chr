package message

// Acknowledgement is transport evidence for the original send.
type Acknowledgement struct {
	MessageID int64
	FileID    string
}
