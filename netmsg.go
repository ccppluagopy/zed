package zed

func (msg *NetMsg) Clone() *NetMsg {
	return &NetMsg{
		Client: msg.Client,
		Cmd:    msg.Cmd,
		Len:    msg.Len,
		Data:   append([]int{}, msg.Data...),
	}
}
