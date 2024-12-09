package server

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/template/html/v2"
	"github.com/gomvn/gomvn/internal/config"
	"github.com/gomvn/gomvn/internal/server/middleware"
	"github.com/gomvn/gomvn/internal/service"
	"github.com/gomvn/gomvn/internal/service/user"
)

func New(conf *config.App, ps *service.PathService, storage *service.Storage, us *user.Service, rs *service.RepoService) *Server {
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		IdleTimeout:           time.Second * 5,
		DisableStartupMessage: true,
		Views:                 engine,
	})

	app.Use(compress.New())

	var cert *Cert
	if conf.Server.Cert != nil {
		cert = &Cert{
			Cert: *conf.Server.Cert,
			Key:  *conf.Server.Key,
		}
	}
	server := &Server{
		app:     app,
		name:    conf.Name,
		listen:  conf.Server.GetListenAddr(),
		cert:    cert,
		ps:      ps,
		storage: storage,
		us:      us,
		rs:      rs,
	}

	api := app.Group("/api")
	api.Use(middleware.NewApiAuth(us))
	api.Get("/users", server.handleApiGetUsers)
	api.Post("/users", server.handleApiPostUsers)
	api.Put("/users/:id", server.handleApiPutUsers)
	api.Delete("/users/:id", server.handleApiDeleteUsers)
	api.Get("/users/:id/refresh", server.handleApiGetUsersRefresh)

	app.Put("/*", middleware.NewRepoAuth(us, ps, true), server.handlePut)
	app.Get("/", server.handleIndex)

	app.Use(middleware.NewRepoAuth(us, ps, false))
	app.Static("/", storage.GetRoot(), fiber.Static{
		Browse: true,
	})

	return server
}

type Cert struct {
	Cert string
	Key  string
}

type Server struct {
	app     *fiber.App
	name    string
	listen  string
	cert    *Cert
	ps      *service.PathService
	storage *service.Storage
	us      *user.Service
	rs      *service.RepoService
}

func (s *Server) Listen() error {
	log.Printf("GoMVN self-hosted repository listening on %s\n", s.listen)
	errCh := make(chan error)
	go func() {
		defer close(errCh)
		err := func() error {
			var err error
			if s.cert != nil {
				err = s.app.ListenTLS(s.listen, s.cert.Cert, s.cert.Key)
				if err != nil {
					return fmt.Errorf("failed to listen: %w", err)
				}
			} else {
				err = s.app.Listen(s.listen)
				if err != nil {
					return fmt.Errorf("failed to listen: %w", err)
				}
			}
			return nil
		}()
		if err != nil {
			errCh <- fmt.Errorf("failed to listen: %w", err)
		}
	}()
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}
	return nil
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
