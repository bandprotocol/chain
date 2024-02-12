package types

// NewIBCChannel creates a new IBCChannel instance.
func NewIBCChannel(portID, channelID string) IBCChannel {
	return IBCChannel{
		PortId:    portID,
		ChannelId: channelID,
	}
}
