// Package coap provides a CoAP client and server.
package coap

import (
	"log"
	"net"
	"time"
)

const maxPktLen = 3000
const timeOut = 30

// Handler is a type that handles CoAP messages.
type Handler interface {
	// Handle the message and optionally return a response message.
	ServeCOAP(l *net.TCPConn, m *Message) *Message
}

type funcHandler func(l *net.TCPConn, m *Message) *Message

func (f funcHandler) ServeCOAP(l *net.TCPConn, m *Message) *Message {
	return f(l, m)
}

// FuncHandler builds a handler from a function.
func FuncHandler(f func(l *net.TCPConn, m *Message) *Message) Handler {
	return funcHandler(f)
}

func handlePacket(l *net.TCPConn, data []byte, rh Handler) {

	msg, err := parseMessage(data)
	if err != nil {
		log.Printf("Error parsing %v", err)
		return
	}

	rv := rh.ServeCOAP(l, &msg)
	if rv != nil {
		Transmit(l, *rv)
	}
}

// Transmit a message.
func Transmit(l *net.TCPConn, m Message) error {
	d, err := m.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = l.Write(d)
	return err
}

// Receive a message.
func Receive(l *net.TCPConn, buf []byte) (Message, error) {
	l.SetReadDeadline(time.Now().Add(ResponseTimeout))
	nr, err := l.Read(buf)
	if err != nil {
		return Message{}, err
	}
	return parseMessage(buf[:nr])
}

// ListenAndServe binds to the given address and serve requests forever.
func ListenAndServe(n, addr string, rh Handler) error {
	taddr, err := net.ResolveTCPAddr(n, addr)
	if err != nil {
		return err
	}

	l, err := net.ListenTCP(n, taddr)
	if err != nil {
		return err
	}
	c, err := l.AcceptTCP()
	if err != nil {
		return err
	}
	// var that = this;
	// this.socket.setNoDelay(true);
	// this.socket.setKeepAlive(true, 15 * 1000); //every 15 second(s)
	// this.socket.on('error', function(err) {
	//     console.log("socket error:", err);
	//     //that.disconnect("socket error " + err);
	// });
	// this.socket.on('close', function(err) {
	//     console.log("socket close:", err);
	//     err ? that.disconnect("socket close " + err) : "";
	// });
	// this.socket.on('timeout', function(err) {
	//     that.disconnect("socket timeout " + err);
	// });

	// this.handshake();
	c.SetNoDelay(true)
	c.SetKeepAlive(true)
	c.SetKeepAlivePeriod(15 * time.Second)
	return Serve(c, rh)
}

func Serve(conn *net.TCPConn, rh Handler) error {
	buf := make([]byte, maxPktLen)
	for {
		nr, err := conn.Read(buf)
		log.Println("Connection from: %q", conn.RemoteAddr().String())
		if err != nil {
			//log.Println("err1: %q", err.Error())
			if neterr, ok := err.(net.Error); ok && (neterr.Temporary() || neterr.Timeout()) {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			return err
		}
		Data := (buf[:nr])
		messnager := make(chan []byte)
		go HeartBeating(conn, messnager, timeOut, rh)
		go GravelChannel(Data, messnager)
	}
}

func HeartBeating(conn *net.TCPConn, readerChannel chan []byte, timeout int, rh Handler) {
	select {
	case fk := <-readerChannel:
		go handlePacket(conn, fk, rh)
		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		break
	case <-time.After(time.Second * 5):
		conn.Close()
	}

}

func GravelChannel(n []byte, mess chan []byte) {
	mess <- n
	close(mess)
}

type SparkCore struct {
}

func startupProtocol() {

}
