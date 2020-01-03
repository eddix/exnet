package exnet

import "errors"

var (
	// ErrNotExnetConn if a net.Conn is not a exnet.Conn
	ErrNotExnetConn = errors.New("Not an ExNet Connection")
	// ErrFreezeExnetConn if a net.Conn is freezed
	ErrFreezeExnetConn = errors.New("Freezed exnet.Conn")
)
