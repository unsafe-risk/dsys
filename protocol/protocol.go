package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"

	"v8.run/go/exp/util/noescape"
	"github.com/unsafe-risk/dsys/encoding2"
	"github.com/unsafe-risk/dsys/encryption"
	"github.com/unsafe-risk/dsys/raft"
)

const (
	protocolVersion = 0x11

	PacketRangeRaftStart = 0x1000
	PacketRangeRaftEnd   = 0x1FFF
)

/*
Header Structure:
=================
1. Version: 1 byte
2. PacketID: 4 bytes
3. body length: 4 bytes
4. body: body length bytes
*/

func WriteFrame(c io.Writer, PacketID uint32, body []byte) error {
	var buffer [9]byte
	buffer[0] = protocolVersion
	binary.BigEndian.PutUint32(buffer[1:5], PacketID)
	binary.BigEndian.PutUint32(buffer[5:9], uint32(len(body)))
	_, err := noescape.Write(c, buffer[:])
	if err != nil {
		return err
	}
	_, err = noescape.Write(c, body)
	return err
}

var ErrInvalidProtocolVersion = errors.New("invalid protocol version")

func ReadHeader(c io.Reader) (PacketID uint32, bodyLength uint32, err error) {
	var buffer [9]byte
	_, err = noescape.Read(c, buffer[:])
	if err != nil {
		return
	}
	if buffer[0] != protocolVersion {
		err = ErrInvalidProtocolVersion
		return
	}
	PacketID = binary.BigEndian.Uint32(buffer[1:5])
	bodyLength = binary.BigEndian.Uint32(buffer[5:9])
	return
}

type Server struct {
	Ln   net.Listener
	Raft *raft.Raft
	Enc  encryption.XChaCha20Poly1305Box
}

func (s *Server) ListenAndServe() error {
	for {
		conn, err := s.Ln.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	var data []byte
	var eb []byte
	var bb bytes.Buffer
	var decrypted []byte
	var enc encryption.XChaCha20Poly1305Box = s.Enc
	for {
		bb.Reset()
		PacketID, bodyLength, err := ReadHeader(conn)
		if err != nil {
			return
		}
		if cap(data) < int(bodyLength) {
			data = make([]byte, bodyLength)
		}
		_, err = io.ReadFull(conn, data[:bodyLength])
		if err != nil {
			return
		}
		decrypted, err = enc.Open(data)
		if err != nil {
			return
		}
		dr := bytes.NewReader(decrypted)
		switch {
		case PacketID >= PacketRangeRaftStart && PacketID <= PacketRangeRaftEnd:
			const (
				PacketTypeHeartbeat    = 0x1001
				PacketTypeHeartbeatACK = 0x1002
				PacketTypeVoteRequest  = 0x1003
				PacketTypeVoteResponse = 0x1004
			)
			switch PacketID {
			case PacketTypeHeartbeat:
				term, _, err := encoding2.LoadUint(dr)
				if err != nil {
					return
				}
				leaderID, err := encoding2.LoadString(dr)
				if err != nil {
					return
				}
				currentTerm, success := s.Raft.Heartbeat(term, leaderID)
				encoding2.Uint(&bb, currentTerm)
				if success {
					encoding2.Uint(&bb, 1)
				} else {
					encoding2.Uint(&bb, 0)
				}
				eb = enc.Seal(eb[:0], bb.Bytes())
				err = WriteFrame(conn, PacketTypeHeartbeatACK, eb)
				if err != nil {
					return
				}
			case PacketTypeVoteRequest:
			}
		}
	}
}
