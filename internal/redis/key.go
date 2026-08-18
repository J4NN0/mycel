package redis

const (
	historyKeyPrefix = "history:"
	activeConvKey    = "conversation:active"
	convSeqKey       = "conversation:seq"
)

func historyKey(conversationID string) string {
	return historyKeyPrefix + conversationID
}
