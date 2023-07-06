package client

import (
	"encoding/binary"
	"magnetron/pkg/proto"
	"net"
)

type UpdateMessage struct {
	MsgType     [2]byte // always has value of 1
	MsgDataSize [2]byte // Remaining size of request
	SrvCount    [2]byte // Number of servers in the server list
	SrvCountDup [2]byte // Same as previous field ¯\_(ツ)_/¯
}

func BuildUpdateMessage(serverMessages []ServerMessage) UpdateMessage {

	serverCount := uint16(len(serverMessages))

	var remainingDataSize int = 0
	for _, staticServer := range serverMessages {
		remainingDataSize += staticServer.GetSizeInBytes()
	}
	remainingDataSize += 4 // 2 bytes for server count, 2 bytes for server count duplicate

	msgDataSize := make([]byte, 2)
	binary.BigEndian.PutUint16(msgDataSize, uint16(remainingDataSize))

	srvCount := make([]byte, 2)
	binary.BigEndian.PutUint16(srvCount, serverCount)
	msg := UpdateMessage{
		MsgType:     [2]byte{0x00, 0x01},
		MsgDataSize: [2]byte{msgDataSize[0], msgDataSize[1]},
		SrvCount:    [2]byte{srvCount[0], srvCount[1]},
		SrvCountDup: [2]byte{srvCount[0], srvCount[1]},
	}

	return msg
}

func SendUpdateMessage(msg UpdateMessage, conn net.Conn) *proto.ProtoError {
	if err := binary.Write(conn, binary.BigEndian, &msg); err != nil {
		result := proto.ProtoError{
			Error:        err,
			ErrorMessage: "Error while attempting to send update message to client.",
			Expected:     msg,
		}
		return &result
	} else {
		return nil
	}
}
