package client

import (
	"encoding/binary"
	"magnetron/pkg/proto"
	"net"
)

type HeaderMessage struct {
	Protocol [4]byte // "HTRK" 0x4854524B
	Version  [2]byte // Old protocol (1) or new (2)
}

func BuildHeaderMessage() HeaderMessage {
	headerMsg := HeaderMessage{
		Protocol: [4]byte{0x48, 0x54, 0x52, 0x4B},
		Version:  [2]byte{0x00, 0x01},
	}

	return headerMsg
}

func ReceiveTrackerHeaderMsg(conn net.Conn) (*HeaderMessage, *proto.ProtoError) {
	var msg = HeaderMessage{}
	if err := binary.Read(conn, binary.BigEndian, &msg); err != nil {
		result := proto.ProtoError{
			Error:        err,
			ErrorMessage: "Error while attempting to receive tracker header message from client)",
			Expected:     msg,
		}
		return nil, &result
	} else {
		return &msg, nil
	}

}

func SendTrackerHeaderMsg(msg HeaderMessage, conn net.Conn) *proto.ProtoError {
	if err := binary.Write(conn, binary.BigEndian, &msg); err != nil {
		result := proto.ProtoError{
			Error:        err,
			ErrorMessage: "Error while attempting to send tracker header message to client",
			Expected:     msg,
		}
		return &result
	} else {
		return nil
	}
}
