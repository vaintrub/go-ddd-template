package client

import (
	"context"
	"net"
	"time"
)

func waitForPort(addr string, timeout time.Duration) bool {
	portAvailable := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		dialer := &net.Dialer{
			Timeout: time.Second,
		}
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// continue
			}

			conn, err := dialer.DialContext(ctx, "tcp", addr)
			if err == nil {
				_ = conn.Close()
				close(portAvailable)
				return
			}

			time.Sleep(time.Millisecond * 200)
		}
	}()

	select {
	case <-portAvailable:
		return true
	case <-ctx.Done():
		return false
	}
}
