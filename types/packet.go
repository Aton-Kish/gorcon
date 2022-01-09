package types

type Packet int32

const (
	AuthRequest     = Packet(3)
	AuthResponse    = Packet(2)
	CommandRequest  = Packet(2)
	CommandResponse = Packet(0)
	DummyRequest    = Packet(100)
)
