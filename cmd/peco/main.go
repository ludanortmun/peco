package main

import (
	"time"

	"github.com/alecthomas/kong"
	"github.com/ludanortmun/peco/internal/core"
	"github.com/ludanortmun/peco/internal/storage"
)

type NewCmd struct {
	Content string     `arg:"" help:"Content of the new todo item."`
	Due     *time.Time `optional:"" help:"Due date of the new todo item." format:"2006-01-02"`
}

func (cmd *NewCmd) Run() error {
	store := getStore()
	defer store.Close()
	client := core.NewPecoClient(store)
	item := core.NewTodo(cmd.Content)
	if cmd.Due != nil {
		item.Due = cmd.Due
	}

	savedItem, err := client.Save(item)
	if err != nil {
		return err
	}

	println("New todo item created with ID:", savedItem.Id)
	return nil
}

var cli struct {
	New NewCmd `cmd:"" help:"Create a new todo item."`
}

func main() {

	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func getStore() storage.SQLiteStore {
	db, err := storage.Open("todos.db")
	if err != nil {
		panic(err)
	}

	return storage.NewSQLiteStore(db)
}
