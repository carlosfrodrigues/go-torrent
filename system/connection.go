package system
import (
	"bytes"
	"fmt"
	"net"
	"time"
	"io"
)

type Client struct {
	Conn net.Conn
	Choked bool
	Bitfield bitfield
	peer Peer
	infoHash []byte
	peerID []byte
}

func NewConnection(peer Peer, peerId string, infoHash []byte)(*Client, error){
	conn, err := net.DialTimeout("tcp", peer.ToString(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	_, err = makeHandshake(conn, peerId, infoHash)
	if err != nil {
		conn.Close()
		return nil, err
	}

	bf, err := getBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn: conn,
		Choked: true,
		Bitfield: bf,
		peer: peer,
		infoHash: infoHash,
		peerID: []byte(peerId),
	}, nil
}

func makeHandshake(conn net.Conn, peerId string, infoHash []byte) ([]byte, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	_, err := conn.Write(generateHandshake(peerId, infoHash))
	if err != nil {
		
		return nil, err
	}

	infoHashReceived, err := readHandshake(conn)

	if !bytes.Equal(infoHashReceived[:], infoHash[:]) {
		return nil, fmt.Errorf("Expected infohash %x but got %x", infoHashReceived, infoHash)
	}
	return infoHashReceived, nil
}

func generateHandshake(peerId string, infoHash []byte) []byte{
	pstr := "BitTorrent protocol"
	hs := make([]byte, len(pstr)+49)
    hs[0] = byte(len(pstr))
    curr := 1
    curr += copy(hs[curr:], pstr)
    curr += copy(hs[curr:], make([]byte, 8))
    curr += copy(hs[curr:], infoHash[:])
	curr += copy(hs[curr:], []byte(peerId[:]))
	return hs
}

func readHandshake(r io.Reader)([]byte, error){
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	pstrlen := int(lengthBuf[0])

	if pstrlen == 0 {
		err := fmt.Errorf("pstrlen cannot be 0")
		return nil, err
	}
	handshakeBuf := make([]byte, 48+pstrlen)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}
	var infoHash []byte
	infoHash = handshakeBuf[pstrlen+8:pstrlen+8+20]
	return infoHash, nil
}

func getBitfield(conn net.Conn) (bitfield, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := ReadMessage(conn)
	if err != nil {
		return nil, err
	}
	if msg.ID != MsgBitfield {
		err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}


func (c *Client) Read() (*Message, error) {
	msg, err := ReadMessage(c.Conn)
	return msg, err
}

func (c *Client) SendRequest(index, begin, length int) error {
	req := FormatRequest(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}

func (c *Client) SendInterested() error {
	msg := Message{ID: MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendNotInterested() error {
	msg := Message{ID: MsgNotInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendUnchoke() error {
	msg := Message{ID: MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendHave(index int) error {
	msg := FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}