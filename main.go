package main

import (
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/ftqo/gothor/api"
	"github.com/ftqo/gothor/config"
	"github.com/ftqo/gothor/db"
	"github.com/ftqo/gothor/logger"
)

func main() {
	// create waitgroup and signal channel
	wg := &sync.WaitGroup{}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	// get config
	conf, err := config.Get()
	if err != nil {
		panic(err)
	}

	// get logger
	log, err := logger.New(conf.Logger)
	if err != nil {
		panic(err)
	}

	// get pool conn pool
	pool, err := db.Open(conf.DB)
	if err != nil {
		log.Panic().Err(err).Send()
	}

	// create session manager
	sm := scs.New()
	sm.Store = postgresstore.New(pool)
	sm.Cookie.SameSite = http.SameSiteStrictMode

	// create api srv
	srv := &api.Server{
		DB:       db.New(pool), // interface with queries
		Sessions: sm,
		Log:      log,
	}

	// start api server
	wg.Add(1)
	go func() {
		err = srv.Start(conf.Server, wg)
		if err != nil {
			pool.Close()
		}
	}()

	// handle shutdown
	<-ch
	go srv.Stop()
	pool.Close()

	wg.Wait()
}
