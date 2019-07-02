package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/daglabs/btcd/rpcclient"
)

type server struct {
	cfg           *config
	rpcConnConfig *rpcclient.ConnConfig

	httpServer *http.Server
	rpcClient  *rpcclient.Client
}

func newServer(cfg *config) (*server, error) {
	server := server{
		cfg: cfg,
	}

	server.rpcConnConfig = &rpcclient.ConnConfig{
		Host:       cfg.Host,
		Endpoint:   "ws",
		User:       cfg.RPCUser,
		Pass:       cfg.RPCPass,
		DisableTLS: cfg.DisableTLS,
	}
	if !cfg.DisableTLS {
		certificate, err := ioutil.ReadFile(cfg.RPCCert)
		if err != nil {
			return nil, err
		}
		server.rpcConnConfig.Certificates = certificate
	}

	return &server, nil
}

func (s *server) start() error {
	log.Printf("Connecting RPC client to %s", s.cfg.Host)

	rpcClient, err := rpcclient.New(s.rpcConnConfig, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to create RPC client: %s", err))
	}
	s.rpcClient = rpcClient

	log.Printf("Starting server on port %d", s.cfg.ListenPort)

	handler := http.NewServeMux()
	handler.HandleFunc("/", s.handleRequest)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.ListenPort),
		Handler: handler,
	}

	return s.httpServer.ListenAndServe()
}

func (s *server) handleRequest(responseWriter http.ResponseWriter, request *http.Request) {
	s.allowCrossOrigin(responseWriter)
	if request.Method == "OPTIONS" {
		// OPTIONS must stop here or else CORS protection will throw a tantrum.
		return
	}

	if request.Method != "POST" {
		responseWriter.WriteHeader(404)
		return
	}

	forwardedResponse, err := s.forwardRequest(request)
	if err != nil {
		responseWriter.WriteHeader(500)
		log.Printf("failed to forward request: %s", err)
		return
	}

	_, err = responseWriter.Write([]byte(forwardedResponse))
	if err != nil {
		responseWriter.WriteHeader(500)
		log.Printf("failed to write response: %s", err)
		return
	}
}

func (s *server) allowCrossOrigin(responseWriter http.ResponseWriter) {
	responseWriter.Header().Set("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	responseWriter.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
}

func (s *server) forwardRequest(request *http.Request) ([]byte, error) {
	jsonRPCMethod := strings.TrimPrefix(request.URL.Path, "/")

	requestBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to read request body: %s", err))
	}

	var jsonRPCParams []json.RawMessage
	err = json.Unmarshal(requestBody, &jsonRPCParams)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to parse params: %s", err))
	}

	response, err := s.rpcClient.RawRequest(jsonRPCMethod, jsonRPCParams)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("request to rpc server failed: %s", err))
	}

	return response, nil
}

func (s *server) stop() error {
	log.Printf("Disconnecting RPC client")
	s.rpcClient.Disconnect()

	log.Printf("Stopping server")
	return s.httpServer.Close()
}
