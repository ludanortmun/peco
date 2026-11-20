package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/alecthomas/kong"
	"github.com/ludanortmun/peco/internal/core"
	"github.com/ludanortmun/peco/internal/mcpapi"
	"github.com/ludanortmun/peco/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "dev"

// localDate interprets terminal input in the system timezone, then passes UTC
// to the domain.
type localDate struct{ time.Time }

func (date *localDate) UnmarshalText(text []byte) error {
	parsed, err := time.ParseInLocation(time.DateOnly, string(text), time.Local)
	if err != nil {
		return err
	}
	date.Time = parsed.UTC()
	return nil
}

type NewCmd struct {
	Content string    `arg:"" help:"Todo content. Quote content containing spaces."`
	Due     localDate `help:"Due date in YYYY-MM-DD format (system local time)." placeholder:"YYYY-MM-DD"`
}

func (cmd *NewCmd) Run(client *core.PecoClient) error {
	item := core.NewTodo(cmd.Content)
	if !cmd.Due.IsZero() {
		item.Due = &cmd.Due.Time
	}
	saved, err := client.Save(item)
	if err != nil {
		return err
	}
	fmt.Printf("Created todo %d\n", saved.Id)
	return nil
}

type ListCmd struct {
	IncludeDone bool `help:"Include completed todos. Pending todos appear first, then earliest due date (undated last), then oldest created."`
}

func (cmd *ListCmd) Run(client *core.PecoClient) error {
	items, err := client.List(cmd.IncludeDone)
	if err != nil {
		return err
	}
	out := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(out, "ID\tSTATE\tDUE\tCONTENT")
	for _, item := range items {
		state, due := "pending", "-"
		if item.State == core.Done {
			state = "done"
		}
		if item.Due != nil {
			due = item.Due.Local().Format(time.DateOnly)
		}
		fmt.Fprintf(out, "%d\t%s\t%s\t%s\n", item.Id, state, due, item.Content)
	}
	return out.Flush()
}

type EditCmd struct {
	ID      uint      `arg:"" help:"ID of the todo to edit."`
	Content *string   `help:"Replace the content. Omit to keep the current content."`
	Due     localDate `help:"Replace the due date, in YYYY-MM-DD format (system local time). Omit to keep the current date." placeholder:"YYYY-MM-DD"`
}

func (cmd *EditCmd) Run(client *core.PecoClient) error {
	item, err := client.Get(cmd.ID)
	if err != nil {
		return err
	}
	if cmd.Content != nil {
		item.Content = *cmd.Content
	}
	if !cmd.Due.IsZero() {
		item.Due = &cmd.Due.Time
	}
	if _, err := client.Save(item); err != nil {
		return err
	}
	fmt.Printf("Updated todo %d\n", item.Id)
	return nil
}

type CompleteCmd struct {
	ID uint `arg:"" help:"ID of the todo to mark as done."`
}

func (cmd *CompleteCmd) Run(client *core.PecoClient) error {
	item, err := client.Get(cmd.ID)
	if err != nil {
		return err
	}
	item.State = core.Done
	if _, err := client.Save(item); err != nil {
		return err
	}
	fmt.Printf("Completed todo %d\n", item.Id)
	return nil
}

type MCPCmd struct{}

func (cmd *MCPCmd) Run(client *core.PecoClient) error {
	return mcpapi.NewServer(client, version).Run(context.Background(), &mcp.StdioTransport{})
}

func main() {
	var cli struct {
		New      NewCmd           `cmd:"" help:"Create a pending todo."`
		List     ListCmd          `cmd:"" help:"List pending todos, sorted by due date and then oldest created."`
		Edit     EditCmd          `cmd:"" help:"Update an existing todo's content or due date."`
		Complete CompleteCmd      `cmd:"" help:"Mark an existing todo as done."`
		MCP      MCPCmd           `cmd:"" help:"Run the MCP server over stdin/stdout."`
		Version  kong.VersionFlag `help:"Print version and exit."`
	}
	ctx := kong.Parse(&cli, kong.Name("peco"), kong.Description("Manage todos stored in your user config directory at peco/peco.db."), kong.Vars{"version": version})
	store, err := storage.OpenConfig()
	ctx.FatalIfErrorf(err)
	defer store.Close()
	err = ctx.Run(core.NewPecoClient(store))
	ctx.FatalIfErrorf(err)
}
