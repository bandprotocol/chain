package types

// NewIBCChannel creates a new IBCChannel instance.
func NewIBCChannel(portId, channelId string) IBCChannel {
	return IBCChannel{
		PortId:    portId,
		ChannelId: channelId,
	}
}
