package registry

import (
	"encoding/binary"
	"fmt"
	"gorm.io/gorm"
	"log"
	"magnetron/internal/config"
	"magnetron/internal/db"
	"magnetron/internal/proto/client"
	"magnetron/internal/proto/server"
	"net"
	"strconv"
	"time"
)

type Registry struct {
	db                    *gorm.DB
	cfg                   *config.Config
	passwordConfig        *config.PasswordConfig
	federatedTrackerStore *db.FederatedTrackerStore
	federatedServerStore  *db.FederatedServerStore
	staticServerStore     *db.StaticServerStore
	registeredServerStore *db.RegisteredServerStore
}

var (
	RegistryInstance *Registry
)

func NewRegistry(cfg *config.Config, passwordConfig *config.PasswordConfig) error {
	var database *gorm.DB
	var err error

	if database, err = db.GetDB(); err != nil {
		return fmt.Errorf("error while getting internal DB connection: %s", err)
	}

	federatedTrackerStore, err := db.NewFederatedTrackerStore(database)

	if err != nil {
		return fmt.Errorf("error while initializing federated tracker store: %s", err)
	}

	federatedServerStore, err := db.NewFederatedServerStore(database)

	if err != nil {
		return fmt.Errorf("error while initializing federated server store: %s", err)
	}

	staticServerStore, err := db.NewStaticServerStore(database)

	if err != nil {
		return fmt.Errorf("error while initializing static server store: %s", err)
	}

	registeredServerStore, err := db.NewRegisteredServerStore(database)

	if err != nil {
		return fmt.Errorf("error while initializing registered server store: %s", err)
	}

	newReg := &Registry{
		db:                    database,
		cfg:                   cfg,
		passwordConfig:        passwordConfig,
		federatedTrackerStore: federatedTrackerStore,
		federatedServerStore:  federatedServerStore,
		staticServerStore:     staticServerStore,
		registeredServerStore: registeredServerStore,
	}

	for idx, entry := range cfg.TrackerFederation.TrackerEntries {
		trackerHost, err := entry.GetHost()
		if err != nil {
			return err
		}

		trackerPort, err := entry.GetPort()
		if err != nil {
			return err
		}

		if _, err := newReg.federatedTrackerStore.RegisterFederatedTracker(trackerHost, trackerPort, entry.Name, entry.Description, entry.UserCount, uint16(idx)); err != nil {
			return err
		}
	}

	for idx, entry := range cfg.StaticEntries {
		var host string
		var port uint16
		order := uint16(idx)

		if host, err = entry.GetHost(); err != nil {
			return err
		}

		if port, err = entry.GetPort(); err != nil {
			return err
		}

		if _, err := newReg.staticServerStore.RegisterStaticServer(host, port, entry.Name, entry.Description, entry.UserCount, order); err != nil {
			return err
		}

	}

	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				newReg.registeredServerStore.RemoveExpiredServers(cfg.ServerExpiration)
			}
		}
	}()

	RegistryInstance = newReg
	return nil

}

func (r *Registry) getAllStaticServers() ([]client.ServerMessage, error) {
	var staticServers []db.StaticServer
	staticServers, err := r.staticServerStore.GetStaticServers()

	if err != nil {
		return nil, err
	}

	serverMessages := make([]client.ServerMessage, 0)

	for _, server := range staticServers {
		if serverMessage, err := client.BuildStaticServerMessage(server); err != nil {
			return nil, err
		} else {
			serverMessages = append(serverMessages, *serverMessage)
		}
	}

	return serverMessages, nil
}

func (r *Registry) getAllRegisteredServers() ([]client.ServerMessage, error) {
	var registeredServers []db.RegisteredServer
	registeredServers, err := r.registeredServerStore.GetAllRegisteredServers()

	if err != nil {
		return nil, err
	}

	serverMessages := make([]client.ServerMessage, 0)

	for _, server := range registeredServers {
		if serverMessage, err := client.BuildRegisteredServerMessage(server); err != nil {
			return nil, err
		} else {
			serverMessages = append(serverMessages, *serverMessage)
		}
	}

	return serverMessages, nil
}

func (r *Registry) getAllFederatedServers() ([]client.ServerMessage, error) {

	serverMessages := make([]client.ServerMessage, 0)

	var federatedTrackers []db.FederatedTracker
	federatedTrackers, err := r.federatedTrackerStore.GetFederatedTrackers()
	if err != nil {
		return nil, err
	}

	for _, tracker := range federatedTrackers {

		if serverMessage, err := client.BuildFederatedTrackerMessage(tracker); err != nil {
			return nil, err
		} else {
			serverMessages = append(serverMessages, *serverMessage)
		}

		federatedServers, err := r.federatedServerStore.GetFederatedServers(tracker.Host, tracker.Port)
		if err != nil {
			return nil, err
		}

		for _, server := range federatedServers {
			if serverMessage, err := client.BuildFederatedServerMessage(server); err != nil {
				return nil, err
			} else {
				serverMessages = append(serverMessages, *serverMessage)
			}
		}
	}

	return serverMessages, nil
}

func (r *Registry) serveClients() {
	server, err := net.Listen("tcp", r.cfg.ClientHost)
	if err != nil {
		log.Fatalln(err)
	}
	defer server.Close()

	log.Println("Tracker is accepting client connections on:", r.cfg.ClientHost)

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println("Failed to accept conn.", err)
			continue
		}

		go func(conn net.Conn) {
			defer func() {
				conn.Close()
			}()
			_, errorMsg := client.ReceiveTrackerHeaderMsg(conn)

			if errorMsg != nil {
				log.Println(errorMsg.ErrorMessage)
			} else {
				log.Printf("Serving hotline client connection from %s", conn.RemoteAddr().String())

				responseHeaderMsg := client.BuildHeaderMessage()
				if msgError := client.SendTrackerHeaderMsg(responseHeaderMsg, conn); msgError != nil {
					log.Println(msgError)
				}

				var staticServerMessages []client.ServerMessage

				if staticServerMessages, err = r.getAllStaticServers(); err != nil {
					log.Println(err)
				}

				var registeredServerMessages []client.ServerMessage

				if registeredServerMessages, err = r.getAllRegisteredServers(); err != nil {
					log.Println(err)
				}

				var federatedServerMessages []client.ServerMessage

				if federatedServerMessages, err = r.getAllFederatedServers(); err != nil {
					log.Println(err)
				}

				var serverMessages []client.ServerMessage

				fedServerHeaderName := []byte(r.cfg.TrackerFederation.Header)

				federatedServerHeader := client.ServerMessage{
					IPAddr:          [4]byte{0, 0, 0, 0},
					Port:            [2]byte{0, 0},
					NumUsers:        [2]byte{0, 0},
					Unused:          [2]byte{0, 0},
					NameSize:        byte(len(fedServerHeaderName)),
					Name:            fedServerHeaderName,
					DescriptionSize: 0,
					Description:     nil,
				}

				serverMessages = append(serverMessages, staticServerMessages...)
				serverMessages = append(serverMessages, registeredServerMessages...)
				serverMessages = append(serverMessages, federatedServerHeader)
				serverMessages = append(serverMessages, federatedServerMessages...)

				update := client.BuildUpdateMessage(serverMessages)

				if msgError := client.SendUpdateMessage(update, conn); msgError != nil {
					log.Println(msgError)
				}

				for _, staticServerMsg := range serverMessages {
					if msgError := client.SendServerRegistry(staticServerMsg, conn); msgError != nil {
						log.Println(msgError)
					}
				}

				client.ReceiveTrackerHeaderMsg(conn)
			}

		}(conn)
	}
}

func (r *Registry) serveServers() {

	log.Println("Tracker is accepting server connections on:", r.cfg.ServerHost)
	hostAddr, err := net.ResolveUDPAddr("udp4", r.cfg.ServerHost)

	conn, err := net.ListenUDP("udp4", hostAddr)
	if err != nil {
		log.Println(err)
	}

	for true {

		block := make([]byte, 2048)
		_, addr, _ := conn.ReadFromUDP(block)

		if serverReg, pError := server.ReadServerRegistration(block); pError != nil {
			log.Println(pError)
		} else {

			passIdArray := make([]byte, 4)
			passIdArray[0] = serverReg.PassId[0]
			passIdArray[1] = serverReg.PassId[1]
			passIdArray[2] = serverReg.PassId[2]
			passIdArray[3] = serverReg.PassId[3]
			passID := binary.BigEndian.Uint32(passIdArray)

			host := addr.IP.String()

			portArray := make([]byte, 2)
			portArray[0] = serverReg.Port[0]
			portArray[1] = serverReg.Port[1]
			port := binary.BigEndian.Uint16(portArray)

			serverName := string(serverReg.Name)

			description := string(serverReg.Description)

			userCountArray := make([]byte, 2)
			userCountArray[0] = serverReg.NumberOfUsers[0]
			userCountArray[1] = serverReg.NumberOfUsers[1]
			userCount := binary.BigEndian.Uint16(userCountArray)

			var validServer = true
			if r.cfg.EnablePasswords == true {

				passwdString := string(serverReg.Password)

				if !CheckPassword(passwdString, *r.passwordConfig) {
					log.Printf("Rejected server %s / %s because of invalid password", serverName, addr.IP.String())
					validServer = false
				}

			}

			if validServer {
				err := r.registeredServerStore.RegisterNewServer(passID, host, port, serverName, description, userCount)
				if err != nil {
					log.Println(err)
				}
			}

			//log.Printf("Registered server %s / %s", serverName, addr.IP.String())

		}
	}
}

func (r *Registry) handleFederatedTrackers() {

	if r.cfg.TrackerFederation.Enabled {

		r.pollFederatedTrackers()

		ticker := time.NewTicker(r.cfg.TrackerFederation.PollFrequency)
		go func() {
			for {
				select {
				case <-ticker.C:
					r.pollFederatedTrackers()
				}
			}
		}()
	}

}

func (r *Registry) pollFederatedTrackers() {
	if trackers, err := r.federatedTrackerStore.GetFederatedTrackers(); err != nil {
		log.Println(err)
	} else {
		for _, tracker := range trackers {
			go r.pollFederatedTracker(tracker.Host, tracker.Port)
		}
	}

}

func (r *Registry) pollFederatedTracker(trackerHost string, trackerPort uint16) {
	conn, err := net.Dial("tcp", trackerHost+":"+strconv.Itoa(int(trackerPort)))

	defer func(conn net.Conn) {
		if conn != nil {
			err := conn.Close()
			if err != nil {
				log.Println(err)
			}
		}
	}(conn)

	if err != nil {
		log.Println(err)
		return
	}

	if msgError := client.SendTrackerHeaderMsg(client.BuildHeaderMessage(), conn); msgError != nil {
		log.Println(msgError)
		return
	}

	if _, errorMsg := client.ReceiveTrackerHeaderMsg(conn); errorMsg != nil {
		log.Println(errorMsg)
		return
	}

	if updateMsg, errorMsg := client.ReceiveUpdateMessage(conn); errorMsg != nil {
		log.Println(errorMsg)
		return
	} else {
		log.Println("Received update message from federated tracker with ", binary.BigEndian.Uint16(updateMsg.SrvCount[:]), " servers")
		for i := 0; i < int(binary.BigEndian.Uint16(updateMsg.SrvCount[:])); i++ {
			if serverMsg, errorMsg := client.ReceiveServerRegistry(conn); errorMsg != nil {
				log.Println(errorMsg)
				return
			} else {

				serverIp := net.IPv4(serverMsg.IPAddr[0], serverMsg.IPAddr[1], serverMsg.IPAddr[2], serverMsg.IPAddr[3]).String()

				portArray := make([]byte, 2)
				portArray[0] = serverMsg.Port[0]
				portArray[1] = serverMsg.Port[1]
				port := binary.BigEndian.Uint16(portArray)

				userCountArray := make([]byte, 2)
				userCountArray[0] = serverMsg.NumUsers[0]
				userCountArray[1] = serverMsg.NumUsers[1]
				userCount := binary.BigEndian.Uint16(userCountArray)

				_, err := r.federatedServerStore.GetFederatedServer(trackerHost, trackerPort, serverIp, port)

				if err == nil {
					if updateError := r.federatedServerStore.UpdateFederatedServer(trackerHost, trackerPort, serverIp, port, string(serverMsg.Name), string(serverMsg.Description), userCount, uint16(i)); updateError != nil {
						log.Println(updateError)
					}
				} else {
					if _, errorMsg := r.federatedServerStore.RegisterFederatedServer(trackerHost, trackerPort, serverIp, port, string(serverMsg.Name), string(serverMsg.Description), userCount, uint16(i)); errorMsg != nil {
						log.Println(errorMsg)
					}
				}
			}
		}
	}
}

func (r *Registry) Serve() {
	go r.serveClients()
	go r.handleFederatedTrackers()
	r.serveServers()
}
