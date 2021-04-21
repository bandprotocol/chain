package types

// NewIBCChannel creates a new IBCChannel instance.
func NewIBCChannel(portId, channelId string) IBCChannel {
	return IBCChannel{
		ChannelId: channelId,
		PortId:    portId,
	}
}
