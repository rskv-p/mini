package bus_leaf

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rskv-p/mini/mod/m_bus/bus_req"
	"github.com/rskv-p/mini/mod/m_bus/bus_sub"
	"github.com/rskv-p/mini/mod/m_bus/bus_type"
)

// Leaf represents a leaf node in the bus system.
type Leaf struct {
	Conn      net.Conn
	Rw        *bufio.ReadWriter
	C         bus_type.IBusClient
	Transform *bus_sub.SubjectTransform
}

func (l *Leaf) CloseConn() error {
	if err := l.Conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}

// initHandlers sets up the necessary handlers for the leaf node.
func (l *Leaf) InitHandlers() {
	// Handle incoming messages
	l.C.SetHandleMessage(func(req *bus_req.Request) {
		subject := req.Subject
		if l.Transform != nil {
			var err error
			subject, err = l.Transform.TransformSubject(subject)
			if err != nil {
				return
			}
		}

		if req.Reply != "" {
			l.SendWithReply(subject, req.Data, req.Reply)
		} else {
			l.Send(subject, req.Data)
		}
	})

	// Handle subscription requests
	l.C.SetOnSubscribe(func(subject string) {
		if l.C.HasMatchingInterest(subject) {
			return
		}

		sub := subject
		if l.Transform != nil {
			var err error
			sub, err = l.Transform.TransformSubject(subject)
			if err != nil {
				return
			}
		}

		l.SendSub(sub)
	})

	// Handle unsubscribe requests
	l.C.SetOnUnsubscribe(func(subject string) {
		sub := subject
		if l.Transform != nil {
			var err error
			sub, err = l.Transform.TransformSubject(subject)
			if err != nil {
				return
			}
		}
		l.SendUnsub(sub)
	})
}

func (l *Leaf) StartPingLoop() {
	go l.pingLoop()
}

func NewLeafNode(remoteAddr string, bus bus_type.IBus, clientFactory bus_type.ClientFactory) (*Leaf, error) {
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial remote address: %w", err)
	}

	client := clientFactory(rand.Uint64(), bus) // Используем фабрику для создания клиента
	leaf := &Leaf{
		Conn: conn,
		Rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		C:    client,
	}

	bus.AddLeaf(leaf) // Теперь leaf реализует интерфейс ILeaf

	infoLine := fmt.Sprintf("INFO {\"id\":%d,\"type\":\"leaf\",\"version\":\"1.0\"}\n", client.GetID())
	_, err = leaf.Rw.WriteString(infoLine)
	if err != nil {
		return nil, fmt.Errorf("failed to send info line: %w", err)
	}
	leaf.Rw.Flush()

	leaf.initHandlers()

	// Start the ping loop in a separate goroutine
	go leaf.pingLoop()

	// Start reading messages in a separate goroutine
	go leaf.ReadLoop()
	return leaf, nil
}

func (l *Leaf) GetConn() net.Conn {
	return l.Conn
}

// pingLoop sends a PING message to the remote server every 30 seconds to keep the connection alive.
func (l *Leaf) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, err := fmt.Fprint(l.Rw, "PING\n")
			if err != nil {
				return
			}
			l.Rw.Flush()
		}
	}
}

// AcceptLeaf handles an incoming leaf connection.
func AcceptLeaf(conn net.Conn, bus bus_type.IBus, clientFactory bus_type.ClientFactory) (*Leaf, error) {
	client := clientFactory(rand.Uint64(), bus) // Используем фабрику для создания клиента
	leaf := &Leaf{
		Conn: conn,
		Rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		C:    client,
	}
	leaf.initHandlers()
	bus.AddLeaf(leaf)

	// Start the ping loop in a separate goroutine
	go leaf.pingLoop()

	// Start reading messages in a separate goroutine
	go leaf.ReadLoop()
	return leaf, nil
}

//---------------------
// Handlers
//---------------------

// initHandlers sets up the necessary handlers for the leaf node.
func (l *Leaf) initHandlers() {
	// Handle incoming messages
	l.C.SetHandleMessage(func(req *bus_req.Request) {
		subject := req.Subject
		if l.Transform != nil {
			var err error
			subject, err = l.Transform.TransformSubject(subject)
			if err != nil {
				return
			}
		}

		if req.Reply != "" {
			l.SendWithReply(subject, req.Data, req.Reply)
		} else {
			l.Send(subject, req.Data)
		}
	})

	// Handle subscription requests
	l.C.SetOnSubscribe(func(subject string) {
		if l.C.HasMatchingInterest(subject) {
			return
		}

		sub := subject
		if l.Transform != nil {
			var err error
			sub, err = l.Transform.TransformSubject(subject)
			if err != nil {
				return
			}
		}

		l.SendSub(sub)
	})

	// Handle unsubscribe requests
	l.C.SetOnUnsubscribe(func(subject string) {
		sub := subject
		if l.Transform != nil {
			var err error
			sub, err = l.Transform.TransformSubject(subject)
			if err != nil {
				return
			}
		}
		l.SendUnsub(sub)
	})
}

//---------------------
// Read Loop
//---------------------

// ReadLoop continuously reads incoming messages and processes commands.
func (l *Leaf) ReadLoop() {
	for {
		line, err := l.Rw.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				// Log error
			}
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "AUTH":
			l.handleAuth(parts)
		case "SUB":
			l.handleSub(parts)
		case "UNSUB":
			l.handleUnsub(parts)
		case "PUB":
			l.handlePublish(parts)
		case "PING":
			l.handlePing()
		case "RESP":
			l.handleResponse(parts)
		default:
			// Log unknown command
		}
	}
}

// handleAuth handles the AUTH command to authenticate the client.
func (l *Leaf) handleAuth(parts []string) {
	if len(parts) < 3 || strings.ToUpper(parts[1]) != "BEARER" {
		l.Rw.WriteString("-ERR invalid auth format\n")
		l.Rw.Flush()
		return
	}
	_, err := verifyJWT(parts[2], l.C.GetSecretKey())
	if err != nil {
		l.Rw.WriteString(fmt.Sprintf("-ERR invalid token: %v\n", err))
		l.Rw.Flush()
		return
	}

	l.Rw.WriteString("+OK\n")
	l.Rw.Flush()
}

// handleSub handles the SUB command to subscribe to a subject.
func (l *Leaf) handleSub(parts []string) {
	if len(parts) > 1 {
		sub := parts[1]
		l.C.MarkRemoteInterest(sub)
		l.C.Subscribe(sub)
	}
}

// handleUnsub handles the UNSUB command to unsubscribe from a subject.
func (l *Leaf) handleUnsub(parts []string) {
	if len(parts) > 1 {
		sub := parts[1]
		l.C.UnmarkRemoteInterest(sub)
		l.C.Unsubscribe(sub)
	}
}

// handlePublish handles the PUB command to publish a message to a subject.
func (l *Leaf) handlePublish(parts []string) {
	if len(parts) > 2 {
		subj := parts[1]
		size, err := strconv.Atoi(parts[2])
		if err != nil {
			l.Rw.WriteString("-ERR invalid size\n")
			l.Rw.Flush()
			return
		}
		msg := make([]byte, size)
		if _, err := io.ReadFull(l.Rw, msg); err != nil {
			l.Rw.WriteString("-ERR read error\n")
			l.Rw.Flush()
			return
		}
		l.C.Publish(subj, msg)
		l.Rw.WriteString("+ACK\n")
		l.Rw.Flush()
	}
}

// handlePing handles the PING command to check if the connection is alive.
func (l *Leaf) handlePing() {
	l.Rw.WriteString("PONG\n")
	l.Rw.Flush()
}

// handleResponse handles the RESP command to process responses.
func (l *Leaf) handleResponse(parts []string) {
	if len(parts) > 2 {
		subj := parts[1]
		size, err := strconv.Atoi(parts[2])
		if err != nil {
			l.Rw.WriteString("-ERR invalid size\n")
			l.Rw.Flush()
			return
		}
		msg := make([]byte, size)
		if _, err := io.ReadFull(l.Rw, msg); err != nil {
			l.Rw.WriteString("-ERR read error\n")
			l.Rw.Flush()
			return
		}
		if l.C.GetHandleMessage() != nil {
			l.C.GetHandleMessage()(&bus_req.Request{
				Subject: subj,
				Data:    msg,
			})
		}
	}
}

//---------------------
// Outgoing Protocol
//---------------------

func (l *Leaf) Send(subject string, msg []byte) {
	_, err := fmt.Fprintf(l.Rw, "PUB %s %d\n", subject, len(msg))
	if err != nil {
		// Log error
	}
	l.Rw.Write(msg)
	l.Rw.Flush()
}

func (l *Leaf) SendWithReply(subject string, msg []byte, reply string) {
	_, err := fmt.Fprintf(l.Rw, "PUB %s %d\n", subject, len(msg))
	if err != nil {
		// Log error
	}
	l.Rw.Write(msg)
	l.Rw.Flush()
}

func (l *Leaf) SendSub(subject string) {
	_, err := fmt.Fprintf(l.Rw, "SUB %s\n", subject)
	if err != nil {
		// Log error
	}
	l.Rw.Flush()
}

func (l *Leaf) SendUnsub(subject string) {
	_, err := fmt.Fprintf(l.Rw, "UNSUB %s\n", subject)
	if err != nil {
		// Log error
	}
	l.Rw.Flush()
}

func (l *Leaf) SendResp(subject string, msg []byte) {
	_, err := fmt.Fprintf(l.Rw, "RESP %s %d\n", subject, len(msg))
	if err != nil {
		// Log error
	}
	l.Rw.Write(msg)
	l.Rw.Write([]byte("\n"))
	l.Rw.Flush()
}

//---------------------
// JWT
//---------------------

func verifyJWT(tokenStr string, secret string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
