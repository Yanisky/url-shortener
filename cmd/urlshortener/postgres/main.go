package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/speps/go-hashids"
	api "github.com/yanisky/url-shortener/api"
	domain "github.com/yanisky/url-shortener/pkg"
	pg "github.com/yanisky/url-shortener/pkg/postgres"
)

func main() {
	var (
		osHashSalt    = os.Getenv("HASH_SALT")
		osServerPort  = os.Getenv("PORT")
		osPostgresURL = os.Getenv("POSTGRES_URL")

		hashSalt    = flag.String("hash_salt", osHashSalt, "Used to salt our id codes")
		serverPort  = flag.String("port", osServerPort, "Http server listening port")
		postgresURL = flag.String("postgres_url", osPostgresURL, "PostgreSQL database url")
	)
	flag.Parse()
	// default port
	addr := ":" + *serverPort
	if len(addr) == 1 {
		addr = ":80"
	}
	fmt.Println(*hashSalt)
	if len(*hashSalt) == 0 {
		panic("No hash salt provided")
	}

	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	// Create id hasher
	hd := hashids.NewData()
	hd.Salt = *hashSalt
	hd.MinLength = 7
	hasher, err := hashids.NewWithData(hd)
	if err != nil {
		panic(err)
	}

	repo, err := pg.NewPostgreSQLRepository(*postgresURL, 60*time.Second, hasher)
	if err != nil {
		panic(err)
	}
	service := domain.NewURLShortenerService(repo, repo)
	server := api.NewGorillaHttpServer()
	handler := api.NewGorillaHTTPHandler(service)

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
