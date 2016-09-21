package coap

import (
	"net"
	"testing"
)

func startTCPLisenter(t *testing.T) (*net.TCPConn, string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("Can't resolve UDP addr")
	}
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		t.Fatal("Can't listen on UDP")
	}
	tcpconn, _ := tcpListener.AcceptTCP()
	coapServerAddr := tcpconn.LocalAddr().String()
	return tcpconn, coapServerAddr
}

func dialAndSend(t *testing.T, addr string, req Message) *Message {
	c, err := Dial("udp", addr)
	if err != nil {
		t.Fatalf("Error dialing: %v", err)
	}
	m, err := c.Send(req)
	if err != nil {
		t.Fatalf("Error sending request: %v", err)
	}
	return m
}

func TestServeWithAckResponse(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      POST,
		MessageID: 9876,
		Payload:   []byte("Content sent by client"),
	}
	req.SetOption(ContentFormat, TextPlain)
	req.SetPathString("/req/path")

	res := Message{
		Type:      Acknowledgement,
		Code:      Content,
		MessageID: req.MessageID,
		Payload:   []byte("Reply from CoAP server"),
	}
	res.SetOption(ContentFormat, TextPlain)
	res.SetPath(req.Path())

	handler := FuncHandler(func(l *net.TCPConn, m *Message) *Message {
		assertEqualMessages(t, req, *m)
		return &res
	})

	udpListener, coapServerAddr := startTCPLisenter(t)
	defer udpListener.Close()
	go Serve(udpListener, handler)

	m := dialAndSend(t, coapServerAddr, req)
	if m == nil {
		t.Fatalf("Didn't receive CoAP response")
	}
	assertEqualMessages(t, res, *m)
}

func TestServeWithoutAckResponse(t *testing.T) {
	req := Message{
		Type:      NonConfirmable,
		Code:      POST,
		MessageID: 54321,
		Payload:   []byte("Content sent by client"),
	}
	req.SetOption(ContentFormat, AppOctets)

	handler := FuncHandler(func(l *net.TCPConn, m *Message) *Message {
		assertEqualMessages(t, req, *m)
		return nil
	})

	udpListener, coapServerAddr := startTCPLisenter(t)
	defer udpListener.Close()
	go Serve(udpListener, handler)

	m := dialAndSend(t, coapServerAddr, req)
	if m != nil {
		t.Fatalf("Received response packet, but expected none")
	}
}
