package zed

type ZServerDelegate interface {
	RecvMsg(*TcpClient) *NetMsg
	SendMsg(*TcpClient) *NetMsg
	HandleMsg(*NetMsg)
}
