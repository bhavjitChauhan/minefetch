//go:build windows

package term

import (
	"context"
	"errors"
	"os"
	"strings"
	"syscall"
	"time"
)

func query(codes ...string) (res string, err error) {
	state, err := makeRaw()
	if err != nil {
		return
	}
	defer setState(syscall.Stdin, state)

	_, err = os.Stdout.WriteString(strings.Join(codes, ""))
	if err != nil {
		return
	}

	ch := make(chan string)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	go func() {
		b := make([]byte, 1, 8)
		n := 0
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			event, err := syscall.WaitForSingleObject(syscall.Stdin, 10)
			if err != nil {
				return
			}
			if event == syscall.WAIT_OBJECT_0 {
				_, err := os.Stdin.Read(b[i : i+1])
				if err != nil {
					return
				}
				b = append(b, 0)
				if b[i] == 't' {
					n++
					if n == len(codes) {
						ch <- string(b)
						return
					}
				}
				i++
			}
		}
	}()
	select {
	case res = <-ch:
	case <-ctx.Done():
		err = errors.New("timeout")
	}
	return
}
