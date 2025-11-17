package tests

import (
	"context"
	"net"
	"time"
)

func WaitForPort(address string) bool {
	waitChan := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
			}

			conn, err := dialer.DialContext(ctx, "tcp", address)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}

			if conn != nil {
				_ = conn.Close()
				waitChan <- struct{}{}
				return
			}
		}
	}()

	select {
	case <-waitChan:
		return true
	case <-ctx.Done():
		return false
	}
}
