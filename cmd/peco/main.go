package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/alecthomas/kong"
	"github.com/ludanortmun/peco/internal/core"
	"github.com/ludanortmun/peco/internal/storage"
)

var version = "dev"

type NewCmd struct {
	Content string    `arg:"" help:"Todo content. Quote content containing spaces."`
	Due     time.Time `help:"Due date in YYYY-MM-DD format (UTC)." format:"2006-01-02"`
}

func (cmd *NewCmd) Run(client *core.PecoClient) error {
	item := core.NewTodo(cmd.Content)
	if !cmd.Due.IsZero() {
		item.Due = &cmd.Due
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
			due = item.Due.Format(time.DateOnly)
		}
		fmt.Fprintf(out, "%d\t%s\t%s\t%s\n", item.Id, state, due, item.Content)
	}
	return out.Flush()
}

type EditCmd struct {
	ID      uint      `arg:"" help:"ID of the todo to edit."`
	Content *string   `help:"Replace the content. Omit to keep the current content."`
	Due     time.Time `help:"Replace the due date, in YYYY-MM-DD format (UTC). Omit to keep the current date." format:"2006-01-02"`
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
		item.Due = &cmd.Due
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

func main() {
	var cli struct {
		New      NewCmd           `cmd:"" help:"Create a pending todo."`
		List     ListCmd          `cmd:"" help:"List pending todos, sorted by due date and then oldest created."`
		Edit     EditCmd          `cmd:"" help:"Update an existing todo's content or due date."`
		Complete CompleteCmd      `cmd:"" help:"Mark an existing todo as done."`
		Version  kong.VersionFlag `help:"Print version and exit."`
	}
	ctx := kong.Parse(&cli, kong.Name("peco"), kong.Description("Manage todos stored in your user config directory at peco/peco.db."), kong.Vars{"version": version})
	store, err := storage.OpenConfig()
	ctx.FatalIfErrorf(err)
	defer store.Close()
	err = ctx.Run(core.NewPecoClient(store))
	ctx.FatalIfErrorf(err)
}
