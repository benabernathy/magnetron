package db

import "testing"

func TestUpdateFederatedServerWritesZeroValues(t *testing.T) {
	database, err := GetDB()
	if err != nil {
		t.Fatal(err)
	}

	store, err := NewFederatedServerStore(database)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.RegisterFederatedServer("10.0.0.2", 5498, "10.0.0.1", 5500, "Server", "Desc", 3, 0x1f95, 1); err != nil {
		t.Fatal(err)
	}

	if err := store.UpdateFederatedServer("10.0.0.2", 5498, "10.0.0.1", 5500, "Server", "Desc", 0, 0, 0); err != nil {
		t.Fatal(err)
	}

	server, err := store.GetFederatedServer("10.0.0.2", 5498, "10.0.0.1", 5500)
	if err != nil {
		t.Fatal(err)
	}

	if server.UserCount != 0 || server.Reserved != 0 || server.ServerOrder != 0 {
		t.Errorf("zero values not persisted: userCount=%d reserved=%d order=%d", server.UserCount, server.Reserved, server.ServerOrder)
	}
}
