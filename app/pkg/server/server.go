package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Server is the main application server for foyle
type Server struct {
	config config.Config
	engine *gin.Engine

	// builtinExtensionPaths is a list of serving paths to the built in extensions
	builtinExtensionPaths []string
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

type staticMapping struct {
	relativePath string
	root         string
}

// createGinEngine sets up the gin engine which is a router
func (s *Server) createGinEngine() error {
	log := zapr.NewLogger(zap.L())
	log.Info("Setting up server")

	router := gin.Default()
	router.GET("/healthz", healthCheck)

	// Serve the static assets for vscode.
	// There should be several directories located in ${ASSETS_DIR}/vscode
	// The second argument to Static is the directory to act as the root for the static files.

	vsCodeRPath := "/out"
	extensionsMapping := staticMapping{
		relativePath: "extensions",
		root:         filepath.Join(s.config.GetAssetsDir(), "vscode/extensions"),
	}

	foyleExtMapping := staticMapping{
		relativePath: "foyle",
		root:         filepath.Join(s.config.GetAssetsDir(), "foyle"),
	}

	mappings := []staticMapping{
		{
			// TODO(jeremy): Can we change "/out" to "/vscode"? We'd have to update various paths in workbench.html
			relativePath: vsCodeRPath,
			root:         filepath.Join(s.config.GetAssetsDir(), "vscode/out-vscode-reh-web-min"),
		},
		{
			relativePath: "resources",
			root:         filepath.Join(s.config.GetAssetsDir(), "vscode/resources"),
		},
		extensionsMapping,
		foyleExtMapping,
	}

	for _, m := range mappings {
		log.Info("Adding vscode static assets", "relativePath", m.relativePath, "root", m.root)
		router.Static(m.relativePath, m.root)
	}

	if err := s.setHTMLTemplates(router); err != nil {
		return err
	}

	// Set the builtin extensions
	if err := s.setVSCodeBuiltinExtensionPaths(extensionsMapping); err != nil {
		return err
	}

	// Add foyle as an extension.
	s.builtinExtensionPaths = append(s.builtinExtensionPaths, foyleExtMapping.relativePath)

	// The workbench endpoint serves the workbench.html page which is the main entry point for vscode for web
	router.GET("/workbench", s.handleGetWorkbench)
	s.engine = router
	return nil
}

// setVSCodeBuiltinExtensionPaths sets the builtin extension paths for the extensions that ship with vscode
func (s *Server) setVSCodeBuiltinExtensionPaths(m staticMapping) error {
	if s.builtinExtensionPaths == nil {
		s.builtinExtensionPaths = make([]string, 0, 30)
	}

	locs, err := findExtensionsInDir(m.root)
	if err != nil {
		return err
	}

	for _, l := range locs {
		relPath, err := filepath.Rel(m.root, l)
		if err != nil {
			return errors.Wrapf(err, "Failed to get relative path for %s relative to %s", l, m.root)
		}
		s.builtinExtensionPaths = append(s.builtinExtensionPaths, filepath.Join(m.relativePath, relPath))
	}
	return nil
}

// handleGetWorkBench handles the request to the workbench endpoint
func (s *Server) handleGetWorkbench(c *gin.Context) {
	// Use the value of the "Host" from the request
	// Extract the origin so that we can make it work behind reverse proxies and the like.
	host := c.Request.Host

	// Lets default the scheme to https but if the host is localhost then assume its http
	scheme := "https"

	if strings.HasPrefix(host, "localhost:") {
		scheme = "http"
	}

	workbenchOpts := IWorkbenchConstructionOptions{
		AdditionalBuiltinExtensions: make([]VSCodeUriComponents, 0, len(s.builtinExtensionPaths)),
	}

	// TODO(jeremy): Should this be cached as a function of scheme and host?
	for _, location := range s.builtinExtensionPaths {
		workbenchOpts.AdditionalBuiltinExtensions = append(workbenchOpts.AdditionalBuiltinExtensions, VSCodeUriComponents{
			Scheme:    scheme,
			Authority: host,
			// Path should be the serving path of the extension
			Path: location,
		})
	}

	opts, err := json.Marshal(workbenchOpts)
	if err != nil {
		if err := c.AbortWithError(http.StatusInternalServerError, err); err != nil {
			// N.B does AbortWithError alway return the error?
			log.Printf("error marshalling workbench options %v", err)
		}
		return
	}

	// Important baseUrl should not have a trailing slash. If it has a trailing slash it will mess up some
	// of the client side code
	baseUrl := scheme + "://" + host
	c.HTML(http.StatusOK, "workbench.html", gin.H{
		"WorkbenchWebBaseUrl":       baseUrl,
		"WorkbenchNLSBaseURL":       baseUrl + "/nls",
		"WorkbenchWebConfiguration": string(opts),
		"WorkbenchAuthSession":      "",
	})
}

// setHTMLTemplates sets the HTML templates for the server.
func (s *Server) setHTMLTemplates(router *gin.Engine) error {
	// Since we are using go:embed to load the templates we can't use the built in
	// gin.LoadHTMLGlob/LoadHTMLFile function. Instead we have to manually parse the templates.
	// This code is based on the code in that file

	// Load the templates we need to explicitly set a name for the template because we aren't using LoadHTMLGlob
	// We don't set delims because we are just using the default delimiters
	templ, err := template.New("workbench.html").Funcs(router.FuncMap).Parse(workbenchTemplateStr)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse workbench template")
	}
	router.SetHTMLTemplate(templ)
	return nil
}

// Run starts the http server
func (s *Server) Run() error {
	address := fmt.Sprintf("%s:%d", s.config.Server.BindAddress, s.config.Server.HttpPort)
	trapInterrupt()
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
