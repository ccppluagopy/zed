package zed

func (msg *NetMsg) Clone() *NetMsg {
	return &NetMsg{
		Client: msg.Client,
		Cmd:    msg.Cmd,
		Data:   msg.Data,
	}
}

func (msg *NetMsg) DeepClone() *NetMsg {
	return &NetMsg{
		Client: msg.Client,
		Cmd:    msg.Cmd,
		Data:   append([]byte{}, msg.Data...),
	}
}
