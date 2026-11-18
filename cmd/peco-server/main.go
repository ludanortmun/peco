package main

import (
	"log"
	"net/http"
	"time"

	"github.com/alecthomas/kong"
	"github.com/ludanortmun/peco/internal/core"
	"github.com/ludanortmun/peco/internal/httpapi"
	"github.com/ludanortmun/peco/internal/storage"
)

var version = "dev"

func main() {
	var cli struct {
		Addr    string           `default:"127.0.0.1:8080" help:"HTTP listen address. Swagger UI is at /docs/."`
		Version kong.VersionFlag `help:"Print version and exit."`
	}
	ctx := kong.Parse(&cli, kong.Name("peco-server"), kong.Description("Serve the todo JSON API using the shared user config database at peco/peco.db."), kong.Vars{"version": version})
	store, err := storage.OpenConfig()
	ctx.FatalIfErrorf(err)
	defer store.Close()
	server := &http.Server{
		Addr:              cli.Addr,
		Handler:           httpapi.New(core.NewPecoClient(store)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("Listening on %s; Swagger UI at /docs/", cli.Addr)
	ctx.FatalIfErrorf(server.ListenAndServe())
}
