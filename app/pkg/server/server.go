package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jlewi/foyle/app/pkg/docs"

	"github.com/jlewi/foyle/protos/go/foyle/logs/logspbconnect"

	"github.com/gin-contrib/cors"

	"connectrpc.com/connect"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/cockroachdb/pebble"

	"github.com/jlewi/foyle/app/pkg/eval"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"

	"github.com/jlewi/foyle/app/pkg/analyze"
	"github.com/jlewi/foyle/app/pkg/logsviewer"
	"github.com/maxence-charriere/go-app/v9/pkg/app"

	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"connectrpc.com/otelconnect"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/agent"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/executor"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Server is the main application server for foyle
type Server struct {
	config  config.Config
	engine  *gin.Engine
	hServer *http.Server
	// builtinExtensionPaths is a list of serving paths to the built in extensions
	builtinExtensionPaths []string

	agent            *agent.Agent
	executor         *executor.Executor
	logsCrud         *analyze.CrudHandler
	sessManager      *analyze.SessionsManager
	shutdownComplete chan bool
}

// NewServer creates a new server
func NewServer(config config.Config, blocksDB *pebble.DB, agent *agent.Agent, tracesDB *pebble.DB, analyzer *analyze.Analyzer, sessManager *analyze.SessionsManager) (*Server, error) {
	e, err := executor.NewExecutor(config)
	if err != nil {
		return nil, err
	}

	if agent == nil {
		return nil, errors.New("Agent is required")
	}

	logsCrud, err := analyze.NewCrudHandler(config, blocksDB, tracesDB, analyzer)
	if err != nil {
		return nil, err
	}
	s := &Server{
		config:      config,
		executor:    e,
		agent:       agent,
		logsCrud:    logsCrud,
		sessManager: sessManager,
	}

	if err := s.createGinEngine(); err != nil {
		return nil, err
	}
	return s, nil
}

type staticMapping struct {
	relativePath string
	root         string
	middleWare   []gin.HandlerFunc
}

// createGinEngine sets up the gin engine which is a router
func (s *Server) createGinEngine() error {
	log := zapr.NewLogger(zap.L())
	log.Info("Setting up server")

	router := gin.Default()

	// TODO(jeremy): Should we turn this into a protobuf service and use connect?
	router.GET("/healthz", s.healthCheck)
	router.NoRoute(func(c *gin.Context) {
		log.Info("Request for not found path", "path", c.Request.URL.Path)
		c.JSON(http.StatusNotFound, gin.H{"message": "Not found", "path": c.Request.URL.Path})
	})

	// TODO(jeremy): We disabled setting up the vscode server because we weren't using it and it requires assets
	// to be setup. Should we get rid of it? Provide a flag to enable it? I do think being able to run vscode
	// in the browser so it is a feature I'd like to add back at some point; see
	// https://github.com/stateful/runme/issues/616
	if false {
		if err := s.serveVSCode(router); err != nil {
			return err
		}
	}

	apiPrefix := s.config.APIPrefix()

	// Set  up the connect-rpc handlers for the EvalServer
	otelInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		return errors.Wrapf(err, "Failed to create otel interceptor")
	}
	path, handler := v1alpha1connect.NewEvalServiceHandler(&eval.EvalServer{}, connect.WithInterceptors(otelInterceptor))
	log.Info("Setting up eval service", "path", path)
	// Since we want to add the prefix apiPrefix we need to strip it before passing it to the connect-rpc handler
	// Refer to https://connectrpc.com/docs/go/routing#prefixing-routes. Note that grpc-go clients don't
	// support prefixes.
	router.Any(apiPrefix+"/"+path+"*any", gin.WrapH(http.StripPrefix("/"+apiPrefix, handler)))

	generatePath, generateHandler := v1alpha1connect.NewGenerateServiceHandler(s, connect.WithInterceptors(otelInterceptor))
	log.Info("Setting up generate service", "path", apiPrefix+"/"+generatePath)
	router.Any(apiPrefix+"/"+generatePath+"*any", gin.WrapH(http.StripPrefix("/"+apiPrefix, generateHandler)))

	aiSvcPath, aiSvcHandler := v1alpha1connect.NewAIServiceHandler(s.agent, connect.WithInterceptors(otelInterceptor))
	log.Info("Setting up AI service", "path", apiPrefix+"/"+aiSvcPath)
	router.Any(apiPrefix+"/"+aiSvcPath+"*any", gin.WrapH(http.StripPrefix("/"+apiPrefix, aiSvcHandler)))

	logsSvcPath, logsSvcHandler := logspbconnect.NewLogsServiceHandler(s.logsCrud, connect.WithInterceptors(otelInterceptor))
	log.Info("Setting up logs service", "path", apiPrefix+"/"+logsSvcPath)
	router.Any(apiPrefix+"/"+logsSvcPath+"*any", gin.WrapH(http.StripPrefix("/"+apiPrefix, logsSvcHandler)))

	sessSvcPath, sessSvcHandler := logspbconnect.NewSessionsServiceHandler(s.sessManager, connect.WithInterceptors(otelInterceptor))
	log.Info("Setting up sessions service", "path", apiPrefix+"/"+sessSvcPath)
	router.Any(apiPrefix+"/"+sessSvcPath+"*any", gin.WrapH(http.StripPrefix("/"+apiPrefix, sessSvcHandler)))

	cSvc, err := docs.NewConvertersService()
	if err != nil {
		return errors.Wrapf(err, "Failed to create ConvertersService")
	}
	cvtSvcPath, cvtSvcHandler := logspbconnect.NewConversionServiceHandler(cSvc, connect.WithInterceptors(otelInterceptor))
	log.Info("Setting up conversion service", "path", apiPrefix+"/"+cvtSvcPath)
	router.Any(apiPrefix+"/"+cvtSvcPath+"*any", gin.WrapH(http.StripPrefix("/"+apiPrefix, cvtSvcHandler)))

	s.engine = router

	// Setup the logs viewer
	if err := s.setupViewerApp(router); err != nil {
		return err
	}
	return nil
}

func (s *Server) serveVSCode(router *gin.Engine) error {
	log := zapr.NewLogger(zap.L())
	// Serve the static assets for vscode.
	// There should be several directories located in ${ASSETS_DIR}/vscode
	// The second argument to Static is the directory to act as the root for the static files.
	vsCodeRPath := "/out"
	extensionsMapping := staticMapping{
		relativePath: extensionsRPath,
		root:         filepath.Join(s.config.GetAssetsDir(), "vscode/extensions"),
		// corsForVSCodeStaticAssets is a hack to deal with
		// https://github.com/microsoft/vscode-discussions/discussions/985.
		// Concretely per that issue, webviews fetch resources from vscode.cdn.net. I probably need to figure out a better
		// long term solution; e.g. do we host those somewhere else? In the interim we configure CORS to allow requests from
		// vscode.cdn.net but only to static assets.
		middleWare: []gin.HandlerFunc{
			cors.New(cors.Config{
				AllowMethods:  []string{"PUT", "PATCH"},
				AllowHeaders:  []string{"Origin"},
				ExposeHeaders: []string{"Content-Length"},
				AllowOriginFunc: func(origin string) bool {
					// Allow requests from vscode-cdn.net
					return strings.HasSuffix(origin, ".vscode-cdn.net")
				},
				AllowCredentials: false,
				MaxAge:           12 * time.Hour,
			}),
		},
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
		group := router.Group(m.relativePath)
		if m.middleWare != nil {
			group.Use(m.middleWare...)
		}
		group.Static("/", m.root)
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

	if s.config.Server.CORS != nil {
		log.Info("Setting up CORS", "config", s.config.Server.CORS)

		for _, origin := range s.config.Server.CORS.AllowedOrigins {
			// We can't allow wildcard or untrusted domains because the server can execute commands via Execute
			// So we rely on CORS to make sure its only coming from a trusted source
			origin := strings.TrimSpace(origin)
			if origin == "*" {
				return errors.New("wildcard is currently not allowed for origins in CORS configuration")
			}
			if origin == "" {
				return errors.New("empty string is currently not allowed for origins in CORS configuration")
			}
		}

		corsConfig := cors.Config{
			AllowOrigins:     s.config.Server.CORS.AllowedOrigins,
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
			AllowHeaders:     s.config.Server.CORS.AllowedHeaders,
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: false,
			MaxAge:           12 * time.Hour,
		}

		if s.config.Server.CORS.VSCodeTestServerPort != nil {
			if *s.config.Server.CORS.VSCodeTestServerPort <= 0 {
				return errors.New("VSCodeTestServerPort must be a positive integer")
			}
			corsFunc, err := NewVscodeCors(*s.config.Server.CORS.VSCodeTestServerPort)
			if err != nil {
				return err
			}
			log.Info("Adding allowed origin for VSCode test server", "port", *s.config.Server.CORS.VSCodeTestServerPort, "regex", corsFunc.match.String())
			corsConfig.AllowOriginFunc = corsFunc.allowOrigin
		}
		corsMiddleWare := cors.New(corsConfig)
		router.Use(corsMiddleWare)
	}
	return nil
}

// setupViewerApp sets up the viewer app
func (s *Server) setupViewerApp(router *gin.Engine) error {
	log := zapr.NewLogger(zap.L())
	app.Route("/", &logsviewer.MainApp{})

	if strings.HasSuffix(logsviewer.AppPath, "/") {
		return errors.New("logsviewer.AppPath should not have a trailing slash")
	}

	if !strings.HasPrefix(logsviewer.AppPath, "/") {
		return errors.New("logsviewer.AppPath should have a leading slash")
	}

	log.Info("Setting up logs viewer", "path", logsviewer.AppPath)

	viewerApp := &app.Handler{
		Name:        "FoyleLogsViewer",
		Description: "View Foyle Logs",
		// Since we don't want to serve the viewer on the root "/" we need to use a CustomProvider
		Resources: app.CustomProvider("", logsviewer.AppPath),
		Styles: []string{
			"/web/viewer.css", // Loads traceSelector.css file.
			"/web/table.css",  // Loads table.css file.
		},
		Env: map[string]string{
			logsviewer.APIPrefixEnvVar: s.config.APIPrefix(),
		},
	}

	// N.B. We need a trailing slash for the relativePath passed to router. Any but not in the stripprefix
	// because we need to leave the final slash in the path so that the route ends up matching.
	router.Any(logsviewer.AppPath+"/*any", gin.WrapH(http.StripPrefix(logsviewer.AppPath, viewerApp)))

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
	s.shutdownComplete = make(chan bool, 1)
	trapInterrupt(s)

	log := zapr.NewLogger(zap.L())
	address := fmt.Sprintf("%s:%d", s.config.Server.BindAddress, s.config.Server.HttpPort)
	log.Info("Starting http server", "address", address)

	hServer := &http.Server{
		WriteTimeout: s.config.Server.HttpMaxWriteTimeout,
		ReadTimeout:  s.config.Server.HttpMaxReadTimeout,
		// We need to wrap it in h2c to support HTTP/2 without TLS
		Handler: h2c.NewHandler(s.engine, &http2.Server{}),
	}
	// Enable HTTP/2 support
	if err := http2.ConfigureServer(hServer, &http2.Server{}); err != nil {
		return errors.Wrapf(err, "failed to configure http2 server")
	}

	s.hServer = hServer

	lis, err := net.Listen("tcp", address)

	if err != nil {
		return errors.Wrapf(err, "Could not start listener")
	}
	go func() {
		if err := hServer.Serve(lis); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error(err, "There was an error with the http server")
			}
		}
	}()

	// Wait for the shutdown to complete
	// We use a channel to signal when the shutdown method has completed and then return.
	// This is necessary because shutdown() is running in a different go function from hServer.Serve. So if we just
	// relied on hServer.Serve to return and then returned from Run we might still be in the middle of calling shutdown.
	// That's because shutdown calls hServer.Shutdown which causes hserver.Serve to return.
	<-s.shutdownComplete
	return nil
}

func (s *Server) shutdown() {
	log := zapr.NewLogger(zap.L())
	log.Info("Shutting down the Foyle server")

	if s.hServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := s.hServer.Shutdown(ctx); err != nil {
			log.Error(err, "Error shutting down http server")
		}
		log.Info("HTTP Server shutdown complete")
	}
	log.Info("Shutdown complete")
	s.shutdownComplete <- true
}

// trapInterrupt shutdowns the server if the appropriate signals are sent
func trapInterrupt(s *Server) {
	log := zapr.NewLogger(zap.L())
	sigs := make(chan os.Signal, 10)
	// Note SIGSTOP and SIGTERM can't be caught
	// We can trap SIGINT which is what ctl-z sends to interrupt the process
	// to interrupt the process
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		msg := <-sigs
		log.Info("Received signal", "signal", msg)
		s.shutdown()
	}()
}

func (s *Server) healthCheck(ctx *gin.Context) {
	// TODO(jeremy): We should return the version
	d := gin.H{
		"server": "foyle",
		"status": "healthy",
	}
	code := http.StatusOK
	ctx.JSON(code, d)
}

// Generate a completion request.
// TODO(https://github.com/jlewi/foyle/issues/173). We should move this function into the Agent structure.
// Its only here on the server because Agent already has a Generate method for the GRPC protocol. We can get
// rid of that method once we get rid of GRPC and GRPCGateway and just use connect.
func (s *Server) Generate(ctx context.Context, req *connect.Request[v1alpha1.GenerateRequest]) (*connect.Response[v1alpha1.GenerateResponse], error) {
	resp, err := s.agent.Generate(ctx, req.Msg)
	cResp := connect.NewResponse(resp)
	return cResp, err
}
