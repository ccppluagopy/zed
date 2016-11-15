package zed

func (msg *NetMsg) Clone() *NetMsg {
	return &NetMsg{
		Client: msg.Client,
		Cmd:    msg.Cmd,
		Len:    msg.Len,
		Data:   msg.Data,
	}
}

func (msg *NetMsg) DeepClone() *NetMsg {
	return &NetMsg{
		Client: msg.Client,
		Cmd:    msg.Cmd,
		Len:    msg.Len,
		Data:   append([]byte{}, msg.Data...),
	}
}
