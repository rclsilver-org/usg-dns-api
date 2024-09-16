package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/sirupsen/logrus"
	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"

	"github.com/rclsilver-org/usg-dns-api/db"
	"github.com/rclsilver-org/usg-dns-api/unifi"
)

type ServerOptions func(opts *config)

func WithVerbose(verbose bool) func(opts *config) {
	return func(opts *config) {
		opts.Verbose = verbose
	}
}

func WithTitle(title string) func(opts *config) {
	return func(opts *config) {
		opts.Title = title
	}
}

func WithVersion(version string) func(opts *config) {
	return func(opts *config) {
		opts.Version = version
	}
}

type Server struct {
	cfg    *config
	db     *db.Database
	unifi  *unifi.Client
	router *fizz.Fizz

	taskTrigger chan bool
}

func NewServer(ctx context.Context, db *db.Database, unifi *unifi.Client, opts ...ServerOptions) (*Server, error) {
	// load the configuration
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to load the configuration: %w", err)
	}
	logrus.WithContext(ctx).Debug("loaded the HTTP server configuration")

	// apply the configuration modifiers
	for _, modifier := range opts {
		modifier(cfg)
	}

	// set the gin release mode when verbose mode is disabled
	if !cfg.Verbose {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.UseRawPath = true

	infos := &openapi.Info{
		Title:   cfg.Title,
		Version: cfg.Version,
	}

	router := fizz.NewFromEngine(engine)
	router.GET("/spec.json", nil, router.OpenAPI(infos, "json"))

	s := &Server{
		cfg:         cfg,
		db:          db,
		router:      router,
		unifi:       unifi,
		taskTrigger: make(chan bool, 1),
	}

	mon := router.Group("/mon", "monitoring", "monitoring of the API")
	{
		mon.GET("/ping", []fizz.OperationOption{
			fizz.Summary("Checks if the API is healthy"),
			fizz.Response(fmt.Sprint(http.StatusInternalServerError), "Server Error", APIError{}, nil, nil),
		}, tonic.Handler(s.monPing, http.StatusOK))
	}

	records := router.Group("/records", "records", "manage the records", s.AuthMiddleware())
	{
		records.GET("", []fizz.OperationOption{
			fizz.Summary("Get the records list"),
			fizz.Response(fmt.Sprint(http.StatusInternalServerError), "Server Error", APIError{}, nil, nil),
		}, tonic.Handler(s.recordList, http.StatusOK))
		records.POST("", []fizz.OperationOption{
			fizz.Summary("Create a new record"),
			fizz.Response(fmt.Sprint(http.StatusInternalServerError), "Server Error", APIError{}, nil, nil),
		}, tonic.Handler(s.recordAdd, http.StatusCreated))
		records.PUT(":record_id", []fizz.OperationOption{
			fizz.Summary("Update an existing record"),
			fizz.Response(fmt.Sprint(http.StatusInternalServerError), "Server Error", APIError{}, nil, nil),
		}, tonic.Handler(s.recordUpdate, http.StatusOK))
		records.DELETE(":record_id", []fizz.OperationOption{
			fizz.Summary("Delete an existing record"),
			fizz.Response(fmt.Sprint(http.StatusInternalServerError), "Server Error", APIError{}, nil, nil),
		}, tonic.Handler(s.recordDelete, http.StatusNoContent))
		records.GET(":record_id", []fizz.OperationOption{
			fizz.Summary("Get an existing record"),
			fizz.Response(fmt.Sprint(http.StatusInternalServerError), "Server Error", APIError{}, nil, nil),
		}, tonic.Handler(s.recordGet, http.StatusOK))
	}

	tonic.SetErrorHook(errorHook)

	return s, nil
}

func (s *Server) RegisterGroup(path, name, description string) *fizz.RouterGroup {
	return s.router.Group(path, name, description)
}

func (s *Server) runTask(ctx context.Context) {
	select {
	case s.taskTrigger <- true:
	default:
		logrus.WithContext(ctx).Debug("task already scheduled, skipping")
	}
}

func (s *Server) StartTaskScheduler(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			var err error

			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				err = s.writeHostsFile(ctx, false)

			case v := <-s.taskTrigger:
				err = s.writeHostsFile(ctx, v)
			}

			if err != nil {
				logrus.WithContext(ctx).WithError(err).Error("unable to write the hosts file")
			}
		}
	}()

	s.taskTrigger <- false
}

func (s *Server) Serve(ctx context.Context) error {
	endpoint := fmt.Sprintf("%s:%d", s.cfg.ListenHost, s.cfg.ListenPort)
	srv := &http.Server{Addr: endpoint, Handler: withLogging(s.router)}

	go func() {
		logrus.WithContext(ctx).Infof("starting the HTTP server on %s", endpoint)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to start the HTTP server")
		}

		logrus.WithContext(ctx).Info("stopped the HTTP server")
	}()

	<-ctx.Done()
	logrus.WithContext(ctx).Info("stopping the HTTP server")

	return srv.Shutdown(context.Background())
}
