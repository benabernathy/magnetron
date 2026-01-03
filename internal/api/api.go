package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"log"
	"magnetron/internal/config"
	"magnetron/internal/db"
	"net/http"
	"strings"
	"time"
)

type RestService struct {
	db                    *gorm.DB
	cfg                   *config.Config
	federatedTrackerStore *db.FederatedTrackerStore
	federatedServerStore  *db.FederatedServerStore
	staticServerStore     *db.StaticServerStore
	registeredServerStore *db.RegisteredServerStore
}

type StaticServersDocument struct {
	Servers []StaticServerDocument `json:"servers"`
}

type StaticServerDocument struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        uint16 `json:"port"`
	Description string `json:"description"`
	UserCount   uint16 `json:"userCount"`
}

type RegisteredServersDocument struct {
	Servers []RegisteredServerDocument `json:"servers"`
}

type RegisteredServerDocument struct {
	Name        string    `json:"name"`
	Host        string    `json:"host"`
	Port        uint16    `json:"port"`
	Description string    `json:"description"`
	UserCount   uint16    `json:"userCount"`
	FirstSeen   time.Time `json:"firstSeen"`
	LastSeen    time.Time `json:"lastSeen"`
}

type FederatedServersDocument struct {
	Servers []FederatedServerDocument `json:"servers"`
}

type FederatedServerDocument struct {
	Name        string    `json:"name"`
	Host        string    `json:"host"`
	Port        uint16    `json:"port"`
	Description string    `json:"description"`
	UserCount   uint16    `json:"userCount"`
	TrackerHost string    `json:"trackerHost"`
	TrackerPort uint16    `json:"trackerPort"`
	FirstSeen   time.Time `json:"firstSeen"`
	LastSeen    time.Time `json:"lastSeen"`
}

type FederatedTrackersDocument struct {
	Trackers []FederatedTrackerDocument `json:"trackers"`
}

type FederatedTrackerDocument struct {
	Name        string    `json:"name"`
	Host        string    `json:"host"`
	Port        uint16    `json:"port"`
	Description string    `json:"description"`
	UserCount   uint16    `json:"userCount"`
	FirstSeen   time.Time `json:"firstSeen"`
	LastSeen    time.Time `json:"lastSeen"`
}

var (
	RestServiceInstance *RestService
)

func NewRestService(cfg *config.Config) error {

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

	if cfg.RestConfig.EnableTokenAuth {
		tokenCfg := ReadTokenConfigFile(cfg.RestConfig.TokenAuthFile)
		tokenCfg.Validate()
	}

	RestServiceInstance = &RestService{
		db:                    database,
		cfg:                   cfg,
		federatedTrackerStore: federatedTrackerStore,
		federatedServerStore:  federatedServerStore,
		staticServerStore:     staticServerStore,
		registeredServerStore: registeredServerStore,
	}

	return nil
}

func (r *RestService) BearerTokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {

		if !r.cfg.RestConfig.EnableTokenAuth {
			next.ServeHTTP(w, request)
		}

		bearerTokenHeaderValue := request.Header.Get("Authorization")

		if bearerTokenHeaderValue == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		authValue := strings.Split(bearerTokenHeaderValue, " ")

		if len(authValue) != 2 {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if authValue[0] != "Bearer" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		suppliedToken := authValue[1]

		tokenCfg := ReadTokenConfigFile(r.cfg.RestConfig.TokenAuthFile)
		tokenCfg.Validate()

		for _, tokenEntry := range tokenCfg.TokenEntries {
			if suppliedToken == tokenEntry.Token {
				if tokenEntry.Expiry.After(time.Now()) {
					ctx := context.WithValue(request.Context(), "token", suppliedToken)
					next.ServeHTTP(w, request.WithContext(ctx))
					return
				}

			}
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}

func (r *RestService) validateBearerToken(request *http.Request) bool {
	bearerTokenHeaderValue := request.Header.Get("Authorization")

	if bearerTokenHeaderValue == "" {
		return false
	}

	authValue := strings.Split(bearerTokenHeaderValue, " ")

	if len(authValue) != 2 {
		return false
	}

	if authValue[0] != "Bearer" {
		return false
	}

	suppliedToken := authValue[1]

	tokenCfg := ReadTokenConfigFile(r.cfg.RestConfig.TokenAuthFile)
	tokenCfg.Validate()

	for _, tokenEntry := range tokenCfg.TokenEntries {
		if suppliedToken == tokenEntry.Token {
			return tokenEntry.Expiry.After(time.Now())
		}
	}

	return false

}

func (r *RestService) getStaticServers(w http.ResponseWriter, request *http.Request) {

	servers, err := r.staticServerStore.GetStaticServers()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	serverDocuments := make([]StaticServerDocument, 0, len(servers))

	for _, server := range servers {
		serverDocument := StaticServerDocument{
			Name:        server.Name,
			Host:        server.Host,
			Port:        server.Port,
			Description: server.Description,
			UserCount:   server.UserCount,
		}

		serverDocuments = append(serverDocuments, serverDocument)
	}

	responseDocument := StaticServersDocument{
		serverDocuments,
	}

	jsonResponse, err := json.Marshal(responseDocument)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)

	if err != nil {
		return
	}

}

func (r *RestService) getRegisteredServers(w http.ResponseWriter, request *http.Request) {

	servers, err := r.registeredServerStore.GetAllRegisteredServers()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	serverDocuments := make([]RegisteredServerDocument, 0, len(servers))

	for _, server := range servers {
		serverDocument := RegisteredServerDocument{
			Name:        server.Name,
			Host:        server.Host,
			Port:        server.Port,
			Description: server.Description,
			UserCount:   server.UserCount,
			FirstSeen:   server.FirstSeen,
			LastSeen:    server.LastSeen,
		}

		serverDocuments = append(serverDocuments, serverDocument)
	}

	responseDocument := RegisteredServersDocument{
		serverDocuments,
	}

	jsonResponse, err := json.Marshal(responseDocument)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
}

func (r *RestService) getFederatedServers(w http.ResponseWriter, request *http.Request) {

	trackers, err := r.federatedTrackerStore.GetFederatedTrackers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	serverDocuments := make([]FederatedServerDocument, 0)

	for _, tracker := range trackers {
		servers, err := r.federatedServerStore.GetFederatedServers(tracker.Host, tracker.Port)

		if err == nil {

			for _, server := range servers {
				serverDocument := FederatedServerDocument{
					Name:        server.Name,
					Host:        server.Host,
					Port:        server.Port,
					TrackerHost: tracker.Host,
					TrackerPort: tracker.Port,
					Description: server.Description,
					UserCount:   server.UserCount,
					FirstSeen:   server.FirstSeen,
					LastSeen:    server.LastSeen,
				}

				serverDocuments = append(serverDocuments, serverDocument)
			}

		}
	}

	responseDocument := FederatedServersDocument{
		serverDocuments,
	}

	jsonResponse, err := json.Marshal(responseDocument)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
}

func (r *RestService) getFederatedTrackers(w http.ResponseWriter, request *http.Request) {

	trackers, err := r.federatedTrackerStore.GetFederatedTrackers()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	trackerDocuments := make([]FederatedTrackerDocument, 0)

	for _, tracker := range trackers {
		trackerDocument := FederatedTrackerDocument{
			Name:        tracker.Name,
			Host:        tracker.Host,
			Port:        tracker.Port,
			Description: tracker.Description,
			UserCount:   tracker.UserCount,
			FirstSeen:   tracker.FirstSeen,
			LastSeen:    tracker.LastSeen,
		}

		trackerDocuments = append(trackerDocuments, trackerDocument)
	}

	responseDocument := FederatedTrackersDocument{
		trackerDocuments,
	}

	jsonResponse, err := json.Marshal(responseDocument)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
}

func (r *RestService) Serve() {

	if !r.cfg.RestConfig.Enabled {
		return
	}

	log.Printf("Serving REST clients from %s", r.cfg.RestConfig.Host)

	if r.cfg.RestConfig.EnableTls {

		http.HandleFunc("GET /api/v1/servers/static/", r.BearerTokenMiddleware(r.getStaticServers))
		http.HandleFunc("GET /api/v1/servers/registered/", r.BearerTokenMiddleware(r.getRegisteredServers))
		http.HandleFunc("GET /api/v1/servers/federated/", r.BearerTokenMiddleware(r.getFederatedServers))
		http.HandleFunc("GET /api/v1/trackers/federated/", r.BearerTokenMiddleware(r.getFederatedTrackers))

		tlsConfig := &tls.Config{
			MinVersion:       tls.VersionTLS12,
			CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			},
		}

		server := &http.Server{
			Addr:         r.cfg.RestConfig.Host,
			Handler:      nil,
			TLSConfig:    tlsConfig,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}

		err := server.ListenAndServeTLS(r.cfg.RestConfig.CertFile, r.cfg.RestConfig.KeyFile)
		if err != nil {
			panic(err)
		}

	} else {

		http.HandleFunc("GET /api/v1/servers/static/", r.getStaticServers)
		http.HandleFunc("GET /api/v1/servers/registered/", r.getRegisteredServers)
		http.HandleFunc("GET /api/v1/servers/federated/", r.getFederatedServers)
		http.HandleFunc("GET /api/v1/trackers/federated/", r.getFederatedTrackers)

		err := http.ListenAndServe(r.cfg.RestConfig.Host, nil)

		if err != nil {
			panic(err)
		}
	}

}
