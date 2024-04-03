package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/config"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Server is the main application server for foyle
type Server struct {
	config config.Config
	engine *gin.Engine
}

// NewServer creates a new server
func NewServer(config config.Config) (*Server, error) {
	s := &Server{
		config: config,
	}

	if err := s.createGinEngine(); err != nil {
		return nil, err
	}
	return s, nil
}

// createGinEngine sets up the gin engine which is a router
func (s *Server) createGinEngine() error {
	log := zapr.NewLogger(zap.L())
	log.Info("Setting up server")

	router := gin.Default()
	router.GET("/healthz", healthCheck)

	s.engine = router
	return nil
}

// Run starts the http server
func (s *Server) Run() error {
	address := fmt.Sprintf("%s:%d", s.config.Server.BindAddress, s.config.Server.HttpPort)
	log.Print("Server listening on http://" + address)
	if err := http.ListenAndServe(address, s.engine); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}

	return nil
}

// trapInterrupt shutdowns the server if the appropriate signals are sent
func trapInterrupt() {
	log := zapr.NewLogger(zap.L())
	sigs := make(chan os.Signal, 10)
	// Note SIGSTOP and SIGTERM can't be caught
	// We can trap SIGINT which is what ctl-z sends to interrupt the process
	// to interrupt the process
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		msg := <-sigs
		log.Info("Received shutdown signal; shutting down the server", "sig", msg)
	}()
}

func healthCheck(ctx *gin.Context) {
	ctx.String(http.StatusOK, "foyle is healthy")
}
