package client

import (
	"encoding/binary"
	"fmt"
	"magnetron/internal/db"
	"magnetron/internal/proto"
	"net"
	"strconv"
	"strings"
)

type ServerMessage struct {
	IPAddr          [4]byte
	Port            [2]byte
	NumUsers        [2]byte // Number of users connected to this particular server
	Unused          [2]byte
	NameSize        byte   // Length of name string
	Name            []byte // Server name
	DescriptionSize byte
	Description     []byte
}

func (msg *ServerMessage) GetSizeInBytes() int {

	var msgSize = 12 // 4 bytes for IP, 2 bytes for port, 2 bytes for num users, 2 bytes for unused,
	// 1 byte for name size, 1 byte for description size

	msgSize = msgSize + len(msg.Description)
	msgSize = msgSize + len(msg.Name)

	return msgSize
}

func (msg *ServerMessage) GetMessageInBytes() []byte {

	var msgBytes []byte

	for _, b := range msg.IPAddr {
		msgBytes = append(msgBytes, b)
	}

	for _, b := range msg.Port {
		msgBytes = append(msgBytes, b)
	}

	for _, b := range msg.NumUsers {
		msgBytes = append(msgBytes, b)
	}

	for _, b := range msg.Unused {
		msgBytes = append(msgBytes, b)
	}

	msgBytes = append(msgBytes, msg.NameSize)

	for _, b := range msg.Name {
		msgBytes = append(msgBytes, b)
	}

	msgBytes = append(msgBytes, msg.DescriptionSize)

	for _, b := range msg.Description {
		msgBytes = append(msgBytes, b)
	}

	return msgBytes
}

func SendServerRegistry(msg ServerMessage, conn net.Conn) *proto.ProtoError {
	// Variable length arrays cannot be automatically converted to byte arrays, we must build the message manually
	msgBytes := msg.GetMessageInBytes()

	if err := binary.Write(conn, binary.BigEndian, &msgBytes); err != nil {
		result := proto.ProtoError{
			Error:        err,
			ErrorMessage: "Error while attempting to send server message to client.",
			Expected:     msg,
		}
		return &result
	} else {
		return nil
	}

}

func ReceiveServerRegistry(conn net.Conn) (*ServerMessage, *proto.ProtoError) {

	ipAddr := make([]byte, 4)
	if err := binary.Read(conn, binary.BigEndian, ipAddr); err != nil {
		result := proto.ProtoError{
			Error:        err,
			ErrorMessage: "Error while attempting to receive server message / ip address from client)",
			Expected:     ipAddr,
		}
		return nil, &result
	}

	port := make([]byte, 2)

	if err := binary.Read(conn, binary.BigEndian, port); err != nil {
		result := proto.ProtoError{
			Error:        err,
			ErrorMessage: "Error while attempting to receive server message / port from client)",
			Expected:     port,
		}
		return nil, &result
	}

	numUsers := make([]byte, 2)

	if err := binary.Read(conn, binary.BigEndian, numUsers); err != nil {
		result := proto.ProtoError{
			Error:        err,
			ErrorMessage: "Error while attempting to receive server message / number of users from client)",
			Expected:     numUsers,
		}
		return nil, &result
	}

	unused := make([]byte, 2)

	if err := binary.Read(conn, binary.BigEndian, unused); err != nil {
		result := proto.ProtoError{
			Error:        err,
			ErrorMessage: "Error while attempting to receive server message / unused from client)",
			Expected:     unused,
		}
		return nil, &result
	}

	nameSize := make([]byte, 1)

	if err := binary.Read(conn, binary.BigEndian, nameSize); err != nil {
		result := proto.ProtoError{
			Error:        err,
			ErrorMessage: "Error while attempting to receive server message / name size from client)",
			Expected:     nameSize,
		}
		return nil, &result
	}

	nameSizeInt := int(nameSize[0])

	name := make([]byte, nameSizeInt)

	if err := binary.Read(conn, binary.BigEndian, name); err != nil {
		result := proto.ProtoError{
			Error:        err,
			ErrorMessage: "Error while attempting to receive server message / name from client)",
			Expected:     name,
		}
		return nil, &result
	}

	descriptionSize := make([]byte, 1)

	if err := binary.Read(conn, binary.BigEndian, descriptionSize); err != nil {
		result := proto.ProtoError{
			Error:        err,
			ErrorMessage: "Error while attempting to receive server message / description size from client)",
			Expected:     descriptionSize,
		}
		return nil, &result
	}

	descriptionSizeInt := int(descriptionSize[0])

	description := make([]byte, descriptionSizeInt)

	if err := binary.Read(conn, binary.BigEndian, description); err != nil {
		result := proto.ProtoError{
			Error:        err,
			ErrorMessage: "Error while attempting to receive server message / description from client)",
			Expected:     description,
		}
		return nil, &result
	}

	msg := ServerMessage{
		IPAddr:          [4]byte{ipAddr[0], ipAddr[1], ipAddr[2], ipAddr[3]},
		Port:            [2]byte{port[0], port[1]},
		NumUsers:        [2]byte{numUsers[0], numUsers[1]},
		Unused:          [2]byte{unused[0], unused[1]},
		NameSize:        nameSize[0],
		Name:            name,
		DescriptionSize: descriptionSize[0],
		Description:     description,
	}

	return &msg, nil
}

func BuildStaticServerMessage(server db.StaticServer) (*ServerMessage, error) {
	ipParts := strings.Split(server.Host, ".")

	if len(ipParts) != 4 {
		return nil, fmt.Errorf("invalid ip address for host - host must be valid ip address: %s", server.Host)
	}

	ipAddr := [4]byte{0x00, 0x00, 0x00, 0x00}

	for index, ipDot := range ipParts {
		if ipValue, err := strconv.Atoi(ipDot); err != nil {
			return nil, err
		} else {
			ipAddr[index] = byte(ipValue)
		}
	}

	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, server.Port)

	numUsers := make([]byte, 2)
	binary.BigEndian.PutUint16(numUsers, server.UserCount)

	serverName := []byte(server.Name)
	serverNameLen := byte(len(serverName))

	description := []byte(server.Description)
	descriptionLen := byte(len(description))

	msg := ServerMessage{
		IPAddr:          ipAddr,
		Port:            [2]byte{port[0], port[1]},
		NumUsers:        [2]byte{numUsers[0], numUsers[1]},
		Unused:          [2]byte{0x00, 0x00},
		NameSize:        serverNameLen,
		Name:            serverName,
		DescriptionSize: descriptionLen,
		Description:     description,
	}

	return &msg, nil
}

func BuildRegisteredServerMessage(server db.RegisteredServer) (*ServerMessage, error) {
	ipParts := strings.Split(server.Host, ".")

	if len(ipParts) != 4 {
		return nil, fmt.Errorf("invalid ip address for host - host must be valid ip address: %s", server.Host)
	}

	ipAddr := [4]byte{0x00, 0x00, 0x00, 0x00}

	for index, ipDot := range ipParts {
		if ipValue, err := strconv.Atoi(ipDot); err != nil {
			return nil, err
		} else {
			ipAddr[index] = byte(ipValue)
		}
	}

	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, server.Port)

	numUsers := make([]byte, 2)
	binary.BigEndian.PutUint16(numUsers, server.UserCount)

	serverName := []byte(server.Name)
	serverNameLen := byte(len(serverName))

	description := []byte(server.Description)
	descriptionLen := byte(len(description))

	msg := ServerMessage{
		IPAddr:          ipAddr,
		Port:            [2]byte{port[0], port[1]},
		NumUsers:        [2]byte{numUsers[0], numUsers[1]},
		Unused:          [2]byte{0x00, 0x00},
		NameSize:        serverNameLen,
		Name:            serverName,
		DescriptionSize: descriptionLen,
		Description:     description,
	}

	return &msg, nil
}

func BuildFederatedServerMessage(server db.FederatedServer) (*ServerMessage, error) {
	ipParts := strings.Split(server.Host, ".")

	if len(ipParts) != 4 {
		return nil, fmt.Errorf("invalid ip address for host - host must be valid ip address: %s", server.Host)
	}

	ipAddr := [4]byte{0x00, 0x00, 0x00, 0x00}

	for index, ipDot := range ipParts {
		if ipValue, err := strconv.Atoi(ipDot); err != nil {
			return nil, err
		} else {
			ipAddr[index] = byte(ipValue)
		}
	}

	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, server.Port)

	numUsers := make([]byte, 2)
	binary.BigEndian.PutUint16(numUsers, server.UserCount)

	serverName := []byte(server.Name)
	serverNameLen := byte(len(serverName))

	description := []byte(server.Description)
	descriptionLen := byte(len(description))

	msg := ServerMessage{
		IPAddr:          ipAddr,
		Port:            [2]byte{port[0], port[1]},
		NumUsers:        [2]byte{numUsers[0], numUsers[1]},
		Unused:          [2]byte{0x00, 0x00},
		NameSize:        serverNameLen,
		Name:            serverName,
		DescriptionSize: descriptionLen,
		Description:     description,
	}

	return &msg, nil
}

func BuildFederatedTrackerMessage(tracker db.FederatedTracker) (*ServerMessage, error) {

	ipParts := strings.Split(tracker.Host, ".")

	ipAddr := [4]byte{0x00, 0x00, 0x00, 0x00}

	if len(ipParts) == 4 {

		for index, ipDot := range ipParts {
			if ipValue, err := strconv.Atoi(ipDot); err != nil {
				return nil, err
			} else {
				ipAddr[index] = byte(ipValue)
			}
		}
	}

	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, tracker.Port)

	numUsers := make([]byte, 2)
	binary.BigEndian.PutUint16(numUsers, tracker.UserCount)

	serverName := []byte(tracker.Name)
	serverNameLen := byte(len(serverName))

	description := []byte(tracker.Description)
	descriptionLen := byte(len(description))

	msg := ServerMessage{
		IPAddr:          ipAddr,
		Port:            [2]byte{port[0], port[1]},
		NumUsers:        [2]byte{numUsers[0], numUsers[1]},
		Unused:          [2]byte{0x00, 0x00},
		NameSize:        serverNameLen,
		Name:            serverName,
		DescriptionSize: descriptionLen,
		Description:     description,
	}

	return &msg, nil
}
