package internet_test

import (
	"context"
	"syscall"
	"testing"

	"github.com/4nd3r5on/Xray-core/common"
	"github.com/4nd3r5on/Xray-core/common/net"
	"github.com/4nd3r5on/Xray-core/testing/servers/tcp"
	. "github.com/4nd3r5on/Xray-core/transport/internet"
)

func TestSockOptMark(t *testing.T) {
	t.Skip("requires CAP_NET_ADMIN")

	tcpServer := tcp.Server{
		MsgProcessor: func(b []byte) []byte {
			return b
		},
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	const mark = 1
	dialer := DefaultSystemDialer{}
	conn, err := dialer.Dial(context.Background(), nil, dest, &SocketConfig{Mark: mark})
	common.Must(err)
	defer conn.Close()

	rawConn, err := conn.(*net.TCPConn).SyscallConn()
	common.Must(err)
	err = rawConn.Control(func(fd uintptr) {
		m, err := syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_MARK)
		common.Must(err)
		if mark != m {
			t.Fatal("unexpected connection mark", m, " want ", mark)
		}
	})
	common.Must(err)
}
