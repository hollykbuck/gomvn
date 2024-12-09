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

	server := &Server{
		app:     app,
		name:    conf.Name,
		listen:  conf.Server.GetListenAddr(),
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

type Server struct {
	app     *fiber.App
	name    string
	listen  string
	ps      *service.PathService
	storage *service.Storage
	us      *user.Service
	rs      *service.RepoService
}

func (s *Server) Listen() error {
	log.Printf("GoMVN self-hosted repository listening on %s\n", s.listen)
	errCh := make(chan error)
	go func() {
		err := s.app.Listen(s.listen)
		if err != nil {
			errCh <- fmt.Errorf("failed to listen: %w", err)
		}
		close(errCh)
	}()
	return <-errCh
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
