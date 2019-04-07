package core

func isTopicMsgChannelClosed(ch chan *TopicMessage) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

