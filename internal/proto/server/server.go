package server

import (
	"errors"
	"magnetron/internal/proto"
)

type ServerRegistration struct {
	magic           [2]byte // Magic number
	Port            [2]byte // Port number
	NumberOfUsers   [2]byte // Number of users connected to this particular server
	Reserved        [2]byte // Undocumented field, relayed to clients as-is
	PassId          [4]byte // Pass ID
	NameSize        byte    // Length of name string
	Name            []byte  // Server name
	DescriptionSize byte    // Length of description string
	Description     []byte  // Server description
	PasswordSize    byte    // Length of password string
	Password        []byte  // Server password
}

func ReadServerRegistration(input []byte) (*ServerRegistration, *proto.ProtoError) {

	var msg ServerRegistration

	if len(input) < 14 {
		return nil, truncatedError(msg)
	}

	msg.magic[0] = input[0]
	msg.magic[1] = input[1]

	msg.Port[0] = input[2]
	msg.Port[1] = input[3]

	msg.NumberOfUsers[0] = input[4]
	msg.NumberOfUsers[1] = input[5]

	msg.Reserved[0] = input[6]
	msg.Reserved[1] = input[7]

	msg.PassId[0] = input[8]
	msg.PassId[1] = input[9]
	msg.PassId[2] = input[10]
	msg.PassId[3] = input[11]

	offset := 12

	msg.NameSize = input[offset]
	offset++

	if offset+int(msg.NameSize)+1 > len(input) {
		return nil, truncatedError(msg)
	}

	msg.Name = input[offset : offset+int(msg.NameSize)]
	offset += int(msg.NameSize)

	msg.DescriptionSize = input[offset]
	offset++

	if offset+int(msg.DescriptionSize) > len(input) {
		return nil, truncatedError(msg)
	}

	msg.Description = input[offset : offset+int(msg.DescriptionSize)]
	offset += int(msg.DescriptionSize)

	// The password block is optional; servers without one may end the message here.
	if offset >= len(input) {
		return &msg, nil
	}

	msg.PasswordSize = input[offset]
	offset++

	if offset+int(msg.PasswordSize) > len(input) {
		return nil, truncatedError(msg)
	}

	msg.Password = input[offset : offset+int(msg.PasswordSize)]

	return &msg, nil

}

func truncatedError(msg ServerRegistration) *proto.ProtoError {
	return &proto.ProtoError{
		Error:        errors.New("message too short"),
		ErrorMessage: "Invalid server registration message",
		Expected:     msg,
	}
}
