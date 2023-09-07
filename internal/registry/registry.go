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
	"time"
)

type Registry struct {
	db             *gorm.DB
	cfg            *config.Config
	passwordConfig *config.PasswordConfig
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
	if err = db.InitDB(database); err != nil {
		return fmt.Errorf("error while initializing internal database: %s", err)
	}

	newReg := &Registry{
		db:             database,
		cfg:            cfg,
		passwordConfig: passwordConfig,
	}

	for _, entry := range cfg.StaticEntries {
		var host string
		var port uint16

		if host, err = entry.GetHost(); err != nil {
			return err
		}

		if port, err = entry.GetPort(); err != nil {
			return err
		}

		if err := newReg.RegisterNewStaticServer(host, port, entry.Name, entry.Description); err != nil {
			return err
		}
	}

	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				removeExpiredServers(newReg)
			}
		}
	}()

	RegistryInstance = newReg
	return nil

}

func (r *Registry) RegisterNewServer(passID uint32, host string, port uint16, name string, description string, userCount uint16) error {

	server := db.RegisteredServer{
		PassID:      passID,
		Host:        host,
		Port:        port,
		Name:        name,
		Description: description,
		UserCount:   userCount,
		LastSeen:    time.Now(),
		FirstSeen:   time.Now(),
	}

	var count int64
	existingServer := &db.RegisteredServer{}
	r.db.First(&existingServer, passID).Count(&count)

	if createError := r.db.Save(&server).Error; createError != nil {
		return fmt.Errorf("could not register server because of an internal error: %s", createError)
	} else if count == 0 {
		log.Printf("Registered new server: %s (%s:%d)", name, host, port)
	}

	return nil
}

func (r *Registry) RegisterNewStaticServer(host string, port uint16, name string, description string) error {

	server := db.StaticServer{
		Host:        host,
		Port:        port,
		Name:        name,
		Description: description,
	}

	if createError := r.db.Create(&server).Error; createError != nil {
		return fmt.Errorf("could not register server because of an internal error: %s", createError)
	}

	return nil

}

func (r *Registry) getStaticServerByName(name string) (db.StaticServer, error) {
	var server db.StaticServer
	if err := r.db.Where("name = ?", name).First(&server).Error; err != nil {
		return db.StaticServer{}, err
	}

	return server, nil

}

func (r *Registry) getRegisteredServerByName(name string) (db.RegisteredServer, error) {
	var server db.RegisteredServer

	if err := r.db.Where("name = ?", name).First(&server).Error; err != nil {
		return db.RegisteredServer{}, err
	}

	return server, nil

}

func (r *Registry) getAllStaticServers() ([]client.ServerMessage, error) {
	var staticServers []db.StaticServer
	r.db.Find(&staticServers)

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
	r.db.Find(&registeredServers)

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

func removeExpiredServers(r *Registry) {

	var registeredServers []db.RegisteredServer

	r.db.Find(&registeredServers)
	//r.db.Raw("SELECT * FROM registered_servers where ").Scan(&registeredServers)

	for _, server := range registeredServers {
		if time.Since(server.LastSeen).Minutes() > r.cfg.ServerExpiration.Minutes() {
			r.db.Delete(&server).Commit()
			log.Printf("Removed expired server %s", server.Name)
		}
	}
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

				var serverMessages []client.ServerMessage

				serverMessages = append(serverMessages, staticServerMessages...)
				serverMessages = append(serverMessages, registeredServerMessages...)

				update := client.BuildUpdateMessage(serverMessages)

				if msgError := client.SendUpdateMessage(update, conn); msgError != nil {
					log.Println(msgError)
				}

				for _, staticServerMsg := range staticServerMessages {
					if msgError := client.SendServerRegistry(staticServerMsg, conn); msgError != nil {
						log.Println(msgError)
					}
				}

				for _, registeredServerMsg := range registeredServerMessages {
					if msgError := client.SendServerRegistry(registeredServerMsg, conn); msgError != nil {
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
				err := r.RegisterNewServer(passID, host, port, serverName, description, userCount)
				if err != nil {
					log.Println(err)
				}
			}

			//log.Printf("Registered server %s / %s", serverName, addr.IP.String())

		}
	}
}

func (r *Registry) Serve() {
	go r.serveClients()
	r.serveServers()
}
