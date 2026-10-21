package server

import (
	"bytes"
	"testing"
)

func buildRegistration(reserved [2]byte, name, description, password string, includePassword bool) []byte {
	msg := []byte{
		0x00, 0x01, // magic
		0x15, 0x7c, // port 5500
		0x00, 0x03, // user count
		reserved[0], reserved[1],
		0xde, 0xad, 0xbe, 0xef, // pass id
	}
	msg = append(msg, byte(len(name)))
	msg = append(msg, name...)
	msg = append(msg, byte(len(description)))
	msg = append(msg, description...)
	if includePassword {
		msg = append(msg, byte(len(password)))
		msg = append(msg, password...)
	}
	return msg
}

func TestReadServerRegistration(t *testing.T) {
	input := buildRegistration([2]byte{0x1f, 0x95}, "My Server", "A description", "secret", true)

	msg, err := ReadServerRegistration(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error)
	}

	if msg.Reserved != [2]byte{0x1f, 0x95} {
		t.Errorf("reserved = %x, want 1f95", msg.Reserved)
	}
	if msg.Port != [2]byte{0x15, 0x7c} {
		t.Errorf("port = %x, want 157c", msg.Port)
	}
	if msg.NumberOfUsers != [2]byte{0x00, 0x03} {
		t.Errorf("user count = %x, want 0003", msg.NumberOfUsers)
	}
	if msg.PassId != [4]byte{0xde, 0xad, 0xbe, 0xef} {
		t.Errorf("pass id = %x, want deadbeef", msg.PassId)
	}
	if string(msg.Name) != "My Server" {
		t.Errorf("name = %q", msg.Name)
	}
	if string(msg.Description) != "A description" {
		t.Errorf("description = %q", msg.Description)
	}
	if string(msg.Password) != "secret" {
		t.Errorf("password = %q", msg.Password)
	}
}

func TestReadServerRegistrationWithoutPassword(t *testing.T) {
	input := buildRegistration([2]byte{0x00, 0x00}, "My Server", "A description", "", false)

	msg, err := ReadServerRegistration(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error)
	}

	if len(msg.Password) != 0 {
		t.Errorf("password = %q, want empty", msg.Password)
	}
}

func TestReadServerRegistrationTruncated(t *testing.T) {
	full := buildRegistration([2]byte{0x1f, 0x95}, "My Server", "A description", "secret", true)

	// Every proper prefix must either parse cleanly or return an error, never panic.
	for i := 0; i < len(full); i++ {
		msg, err := ReadServerRegistration(full[:i])
		if err == nil && msg == nil {
			t.Errorf("prefix of length %d returned neither message nor error", i)
		}
	}

	// Length fields that point past the end of the message must be rejected.
	cases := map[string][]byte{
		"name too long":        append(bytes.Clone(full[:12]), 0xff, 'a'),
		"description too long": append(bytes.Clone(full[:12]), 0x01, 'a', 0xff, 'b'),
		"password too long":    append(bytes.Clone(full[:12]), 0x01, 'a', 0x01, 'b', 0xff, 'c'),
	}

	for name, input := range cases {
		if _, err := ReadServerRegistration(input); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}
