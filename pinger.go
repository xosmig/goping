package goping

import (
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"math"
	"net"
	"os"
	"time"
)

type UnexpectedError string
func (err UnexpectedError) Error() string {
	return "something went wrong: " + string(err)
}

type BadReply string
func (err BadReply) Error() string {
	return string(err)
}

const (
	maxICMPPacketSize int = math.MaxUint16 + 1
	// See en.wikipedia.org/wiki/List_of_IP_protocol_numbers
	ipv4ICMPProtocolNumber = 1
	// Listen all IPv4 addresses on the local machine.
	// See en.wikipedia.org/wiki/0.0.0.0
	listenAddress = "0.0.0.0"
)

var seq uint16 = 0

func nextSeq() int {
	res := seq
	if seq == math.MaxUint16 {
		seq = 0
	} else {
		seq++
	}
	return int(res)
}

func PingOnce(destination *net.IPAddr) {
	PingOnceWithTimeout(destination, -1)
}

func PingOnceWithTimeout(destination *net.IPAddr, timeout time.Duration) error {
	// Create ICMP message
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0, // ICMP subtype, see https://en.wikipedia.org/wiki/Internet_Control_Message_Protocol#Control_messages
		Body: &icmp.Echo{
			ID:  os.Getpid(),
			Seq: nextSeq(),
		},
	}
	binary_msg, err := msg.Marshal([]byte{})
	if err != nil { return err }

	// Create connection
	connection, err := icmp.ListenPacket("ip4:icmp", listenAddress)
	if err != nil { return err }
	defer connection.Close()

	// Send the message
	if n, err := connection.WriteTo(binary_msg, destination); err != nil {
		return err
	} else if n != len(binary_msg) {
		return UnexpectedError(fmt.Sprintf("expexted to send %v bytes, but sent %v", len(binary_msg), n))
	}

	// Get a reply
	if timeout > 0 {
		if err := connection.SetReadDeadline(time.Now().Add(timeout)); err != nil { return err }
	}
	binary_reply := make([]byte, maxICMPPacketSize)
	if _, _, err := connection.ReadFrom(binary_reply); err != nil { return err }

	reply, err := icmp.ParseMessage(ipv4ICMPProtocolNumber, binary_reply)
	if err != nil { return err }

	switch reply.Type {
	case ipv4.ICMPTypeEchoReply:
		return nil
	case ipv4.ICMPTypeTimeExceeded:
		return BadReply("Timeout exceeded.")
	default:
		return UnexpectedError(fmt.Sprintf("Unexpected reply type: %v", reply))
	}
}
