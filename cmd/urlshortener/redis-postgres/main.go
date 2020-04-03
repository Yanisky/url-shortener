package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/speps/go-hashids"
	api "github.com/yanisky/url-shortener/api"
	domain "github.com/yanisky/url-shortener/pkg"
	pg "github.com/yanisky/url-shortener/pkg/postgres"
	redis "github.com/yanisky/url-shortener/pkg/redis"
)

type Server struct {
	Router *mux.Router
}

func NewGorillaHttpServer() Server {
	return Server{
		Router: mux.NewRouter(),
	}
}
func (s *Server) Route(handler api.URLShortnerHttpHandler) {
	s.Router.HandleFunc("/{urlHash}", handler.Redirect).Methods("GET")
	s.Router.HandleFunc("/api/v1/urls", handler.CreateURL).Methods("POST")
	s.Router.HandleFunc("/api/v1/urls/{urlHash}/views", handler.ViewUrlStats).Methods("GET")
}
func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.Router)
}

func main() {
	var (
		osHashSalt    = os.Getenv("HASH_SALT")
		osServerPort  = os.Getenv("PORT")
		osPostgresURL = os.Getenv("POSTGRES_URL")
		osRedisURL    = os.Getenv("REDIS_URL")

		hashSalt    = flag.String("hash_salt", osHashSalt, "Used to salt our id codes")
		serverPort  = flag.String("port", osServerPort, "Http server listening port")
		postgresURL = flag.String("postgres_url", osPostgresURL, "PostgreSQL database url")
		redisURL    = flag.String("redis_url", osRedisURL, "Redis url")
	)
	flag.Parse()
	// default port
	addr := ":" + *serverPort
	if len(addr) == 1 {
		addr = ":80"
	}
	if len(*hashSalt) == 0 {
		panic("No hash salt provided")
	}
	// logger
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	// id "hasher"
	hd := hashids.NewData()
	hd.Salt = *hashSalt
	hd.MinLength = 7
	hasher, _ := hashids.NewWithData(hd)

	postgresRepo, err := pg.NewPostgreSQLRepository(*postgresURL, 60*time.Second, hasher)
	if err != nil {
		panic(err)
	}

	redisCache, err := redis.NewRedisRepository(*redisURL, 60*time.Second, hasher)
	if err != nil {
		panic(err)
	}

	// service - no cache - postgresRepo implements both URL
	simpleService := domain.NewURLShortenerService(postgresRepo, postgresRepo)
	// wrap service with cache
	cachedService := domain.NewCachedURLShortenerService(simpleService, redisCache)

	server := api.NewGorillaHttpServer()
	handler := api.NewGorillaHTTPHandler(cachedService)

	server.Route(handler)

	errChan := make(chan error, 2)

	go func() {
		logger.Log("transport", "http", "address", addr, "msg", "listening")
		errChan <- server.Run(addr)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errChan)

}
