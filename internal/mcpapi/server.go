// Package mcpapi exposes peco's domain client as MCP tools.
package mcpapi

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ludanortmun/peco/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listInput struct {
	IncludeDone bool `json:"include_done,omitempty" jsonschema:"Include completed TODOs (default false)."`
}

type createInput struct {
	Content string  `json:"content" jsonschema:"TODO content."`
	Due     *string `json:"due,omitempty" jsonschema:"Optional due date in local YYYY-MM-DD format."`
}

type editInput struct {
	TodoID  uint    `json:"todo_id" jsonschema:"The existing peco TODO ID."`
	Content *string `json:"content,omitempty" jsonschema:"Replacement content; omit or pass null to preserve."`
	Due     *string `json:"due,omitempty" jsonschema:"Replacement local YYYY-MM-DD date; omit or pass null to preserve."`
}

type completeInput struct {
	TodoID uint `json:"todo_id" jsonschema:"The existing peco TODO ID."`
}

// NewServer registers tools that call the shared core domain logic.
func NewServer(client *core.PecoClient, version string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "peco", Version: version}, nil)
	api := &api{client: client}
	mcp.AddTool(server, &mcp.Tool{Name: "list_todos", Description: "List peco TODOs with IDs, state, due dates and content; optionally include done."}, api.list)
	mcp.AddTool(server, &mcp.Tool{Name: "create_todo", Description: "Create a pending TODO. Optional due date uses local YYYY-MM-DD."}, api.create)
	mcp.AddTool(server, &mcp.Tool{Name: "edit_todo", Description: "Edit an existing TODO by ID, preserving fields whose arguments are omitted or null."}, api.edit)
	mcp.AddTool(server, &mcp.Tool{Name: "complete_todo", Description: "Mark an existing TODO as done using its peco ID."}, api.complete)
	return server
}

type api struct{ client *core.PecoClient }

func textResult(text string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}, nil, nil
}

func parseDue(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	due, err := time.ParseInLocation(time.DateOnly, *value, time.Local)
	if err != nil {
		return nil, fmt.Errorf("due must be a valid local YYYY-MM-DD date: %w", err)
	}
	due = due.UTC()
	return &due, nil
}

func (api *api) list(_ context.Context, _ *mcp.CallToolRequest, input listInput) (*mcp.CallToolResult, any, error) {
	items, err := api.client.List(input.IncludeDone)
	if err != nil {
		return nil, nil, err
	}
	var buf strings.Builder
	out := tabwriter.NewWriter(&buf, 0, 4, 2, ' ', 0)
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
	if err := out.Flush(); err != nil {
		return nil, nil, err
	}
	return textResult(buf.String())
}

func (api *api) create(_ context.Context, _ *mcp.CallToolRequest, input createInput) (*mcp.CallToolResult, any, error) {
	due, err := parseDue(input.Due)
	if err != nil {
		return nil, nil, err
	}
	item := core.NewTodo(input.Content)
	item.Due = due
	saved, err := api.client.Save(item)
	if err != nil {
		return nil, nil, err
	}
	return textResult(fmt.Sprintf("Created todo %d", saved.Id))
}

func (api *api) edit(_ context.Context, _ *mcp.CallToolRequest, input editInput) (*mcp.CallToolResult, any, error) {
	item, err := api.client.Get(input.TodoID)
	if err != nil {
		return nil, nil, err
	}
	if input.Content != nil {
		item.Content = *input.Content
	}
	if input.Due != nil {
		item.Due, err = parseDue(input.Due)
		if err != nil {
			return nil, nil, err
		}
	}
	if _, err := api.client.Save(item); err != nil {
		return nil, nil, err
	}
	return textResult(fmt.Sprintf("Updated todo %d", item.Id))
}

func (api *api) complete(_ context.Context, _ *mcp.CallToolRequest, input completeInput) (*mcp.CallToolResult, any, error) {
	item, err := api.client.Get(input.TodoID)
	if err != nil {
		return nil, nil, err
	}
	item.State = core.Done
	if _, err := api.client.Save(item); err != nil {
		return nil, nil, err
	}
	return textResult(fmt.Sprintf("Completed todo %d", item.Id))
}
