package server

import (
	"errors"
	"magnetron/pkg/proto"
)

type ServerRegistration struct {
	magic           [2]byte // Magic number
	Port            [2]byte // Port number
	NumberOfUsers   [2]byte // Number of users connected to this particular server
	magic2          [2]byte // Magic number part deux
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

	if len(input) < 17 {
		result := proto.ProtoError{
			Error:        errors.New("message too short"),
			ErrorMessage: "Invalid server registration message",
			Expected:     msg,
		}
		return nil, &result
	}

	msg.magic[0] = input[0]
	msg.magic[1] = input[1]

	msg.Port[0] = input[2]
	msg.Port[1] = input[3]

	msg.NumberOfUsers[0] = input[4]
	msg.NumberOfUsers[1] = input[5]

	msg.magic2[0] = input[6]
	msg.magic2[1] = input[7]

	msg.PassId[0] = input[8]
	msg.PassId[1] = input[9]
	msg.PassId[2] = input[10]
	msg.PassId[3] = input[11]

	msg.NameSize = input[12]

	msg.Name = input[13 : 13+msg.NameSize]

	msg.DescriptionSize = input[13+msg.NameSize]

	msg.Description = input[14+msg.NameSize : 14+msg.NameSize+msg.DescriptionSize]

	msg.PasswordSize = input[14+msg.NameSize+msg.DescriptionSize]

	msg.Password = input[15+msg.NameSize+msg.DescriptionSize : 15+msg.NameSize+msg.DescriptionSize+msg.PasswordSize]

	return &msg, nil

}
