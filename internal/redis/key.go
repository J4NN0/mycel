package redis

import "fmt"

const (
	historyKeyPrefix    = "history:"
	activeConvKeyPrefix = "conversation:active:"
	convSeqKeyPrefix    = "conversation:seq:"
)

func historyKey(sessionID, conversationID string) string {
	return fmt.Sprintf("%s%s:%s", historyKeyPrefix, sessionID, conversationID)
}

func activeConvKey(sessionID string) string {
	return activeConvKeyPrefix + sessionID
}

func convSeqKey(sessionID string) string {
	return convSeqKeyPrefix + sessionID
}
