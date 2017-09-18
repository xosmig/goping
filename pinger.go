package goping

import (
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"math"
	"net"
	"os"
	"time"
	"io"
	"sync"
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

var seq int32 = 0
var mutex = &sync.Mutex{}

func nextSeq() int {
	mutex.Lock()
	defer mutex.Unlock()
	res := int(seq)
	if seq == math.MaxUint16 {
		seq = 0
	} else {
		seq++
	}
	return res
}

func PingOnce(destination *net.IPAddr, timeout time.Duration) error {
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

	// Set deadline
	if timeout > 0 {
		if err := connection.SetDeadline(time.Now().Add(timeout)); err != nil { return err }
	}

	// Send the message
	if n, err := connection.WriteTo(binary_msg, destination); err != nil {
		return err
	} else if n != len(binary_msg) {
		return UnexpectedError(fmt.Sprintf("expexted to send %v bytes, but sent %v", len(binary_msg), n))
	}

	// Get a reply
	binary_reply := make([]byte, maxICMPPacketSize)
	if _, _, err := connection.ReadFrom(binary_reply); err != nil { return err }

	reply, err := icmp.ParseMessage(ipv4ICMPProtocolNumber, binary_reply)
	if err != nil { return err }

	if reply.Type != ipv4.ICMPTypeEchoReply {
		return UnexpectedError(fmt.Sprintf("Unexpected reply type: %v", reply.Type))
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	} else  {
		return b
	}
}

func UrlReachable(params Params, output io.Writer) bool {
	dst, err := net.ResolveIPAddr("ip", params.Url)
	if err != nil {
		fmt.Fprintln(output, err.Error())
		return false
	}

	timeout := time.Second * time.Duration(params.Timeout)
	deadline := time.Now().Add(time.Second * time.Duration(params.Deadline))
	// infinity if params.Count < 0
	for i := 0; i != params.Count; i++ {
		fmt.Fprintf(output, "Ping %v (%v)\n", params.Url, dst)
		err := PingOnce(dst, time.Second * time.Duration(params.Timeout))
		if err != nil {
			fmt.Fprintln(output, err.Error())
			if beforeDeadline := deadline.Sub(time.Now()); params.Interval == -1 || beforeDeadline < timeout {
				PingOnce(dst, beforeDeadline)
			} else {
				PingOnce(dst, timeout)
			}

			if !time.Now().Before(deadline) {
				break
			}
			time.Sleep(time.Duration(max(params.Interval - params.Timeout, 0)) * time.Second)
		} else {
			fmt.Fprintf(output, "Url %v (%v) is reachable\n", params.Url, dst.IP)
			return true
		}
	}

	fmt.Fprintf(output, "Url %v (%v) is not reachable\n", params.Url, dst.IP)
	return false
}
