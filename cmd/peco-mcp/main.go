package main

import (
	"context"

	"github.com/alecthomas/kong"
	"github.com/ludanortmun/peco/internal/core"
	"github.com/ludanortmun/peco/internal/mcpapi"
	"github.com/ludanortmun/peco/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "dev"

func main() {
	var cli struct {
		Version kong.VersionFlag `help:"Print version and exit."`
	}
	ctx := kong.Parse(&cli, kong.Name("peco-mcp"), kong.Description("Serve the todo MCP tools over stdin/stdout using the shared user config database at peco/peco.db."), kong.Vars{"version": version})
	store, err := storage.OpenConfig()
	ctx.FatalIfErrorf(err)
	defer store.Close()
	err = mcpapi.NewServer(core.NewPecoClient(store), version).Run(context.Background(), &mcp.StdioTransport{})
	ctx.FatalIfErrorf(err)
}
