package client

import (
	"bytes"
	"magnetron/internal/db"
	"net"
	"testing"
)

func TestServerMessageRoundTripPreservesReserved(t *testing.T) {
	sent := ServerMessage{
		IPAddr:          [4]byte{192, 168, 1, 10},
		Port:            [2]byte{0x15, 0x7c},
		NumUsers:        [2]byte{0x00, 0x07},
		Reserved:        [2]byte{0x1f, 0x95},
		NameSize:        9,
		Name:            []byte("My Server"),
		DescriptionSize: 13,
		Description:     []byte("A description"),
	}

	if got := len(sent.GetMessageInBytes()); got != sent.GetSizeInBytes() {
		t.Fatalf("encoded length = %d, GetSizeInBytes = %d", got, sent.GetSizeInBytes())
	}

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	go func() {
		if err := SendServerRegistry(sent, serverConn); err != nil {
			t.Errorf("send failed: %v", err.Error)
		}
	}()

	received, err := ReceiveServerRegistry(clientConn)
	if err != nil {
		t.Fatalf("receive failed: %v", err.Error)
	}

	if received.Reserved != sent.Reserved {
		t.Errorf("reserved = %x, want %x", received.Reserved, sent.Reserved)
	}
	if received.IPAddr != sent.IPAddr || received.Port != sent.Port || received.NumUsers != sent.NumUsers {
		t.Errorf("header fields mismatch: got %+v", received)
	}
	if !bytes.Equal(received.Name, sent.Name) || !bytes.Equal(received.Description, sent.Description) {
		t.Errorf("name/description mismatch: got %q / %q", received.Name, received.Description)
	}
}

func TestBuildRegisteredServerMessageRelaysReserved(t *testing.T) {
	msg, err := BuildRegisteredServerMessage(db.RegisteredServer{
		Host:     "10.0.0.1",
		Port:     5500,
		Name:     "My Server",
		Reserved: 0x1f95,
	})
	if err != nil {
		t.Fatal(err)
	}

	if msg.Reserved != [2]byte{0x1f, 0x95} {
		t.Errorf("reserved = %x, want 1f95", msg.Reserved)
	}
}

func TestBuildFederatedServerMessageRelaysReserved(t *testing.T) {
	msg, err := BuildFederatedServerMessage(db.FederatedServer{
		FederatedServerId: db.FederatedServerId{Host: "10.0.0.1", Port: 5500},
		Name:              "My Server",
		Reserved:          0x1f95,
	})
	if err != nil {
		t.Fatal(err)
	}

	if msg.Reserved != [2]byte{0x1f, 0x95} {
		t.Errorf("reserved = %x, want 1f95", msg.Reserved)
	}
}
