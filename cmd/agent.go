package cmd

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/smitt14ua/zeus/internal/agent"
	"github.com/smitt14ua/zeus/internal/protocol"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run zeus as a remote-controlled agent connected to a web panel",
	Example: "  zeus agent --url wss://panel.example.com/ws/agent --token abc123\n" +
		"  zeus agent --url wss://panel.example.com/ws/agent --token abc123 --name server-A",
	Args: cobra.NoArgs,
	Run:  runAgent,
}

func runAgent(cmd *cobra.Command, args []string) {
	url, _ := cmd.Flags().GetString("url")
	token, _ := cmd.Flags().GetString("token")
	name, _ := cmd.Flags().GetString("name")
	heartbeat, _ := cmd.Flags().GetDuration("heartbeat")
	reconnect, _ := cmd.Flags().GetDuration("reconnect")

	if url == "" {
		fatalf("--url is required")
	}
	if token == "" {
		fatalf("--token is required")
	}
	if name == "" {
		h, err := os.Hostname()
		if err != nil {
			fatalf("resolving hostname: %v", err)
		}
		name = h
	}

	exec := buildExecutor()

	client := &agent.Client{
		URL:               url,
		Token:             token,
		Name:              name,
		Version:           version,
		HeartbeatInterval: heartbeat,
		ReconnectDelay:    reconnect,
		Executor:          exec,
	}

	if err := client.Run(cmd.Context()); err != nil && err != context.Canceled {
		fatal(err)
	}
}

// buildExecutor registers all supported remote commands.
func buildExecutor() *agent.Executor {
	exec := agent.NewExecutor()

	exec.Register(protocol.CmdProfileNew, func(ctx context.Context, args map[string]any, w io.Writer) error {
		return execProfileNew(agent.StringArg(args, "name"), agent.StringArg(args, "format"), w)
	})

	exec.Register(protocol.CmdProfileAdd, func(ctx context.Context, args map[string]any, w io.Writer) error {
		content := agent.StringArg(args, "content")
		format := agent.StringArg(args, "format")
		if format == "" {
			format = "json"
		}
		return execProfileAdd(ctx, strings.NewReader(content), format,
			agent.StringArg(args, "name"), agent.BoolArg(args, "copy_keys"), true, w)
	})

	exec.Register(protocol.CmdProfileList, func(ctx context.Context, args map[string]any, w io.Writer) error {
		return execProfileList(ctx, "json", w)
	})

	exec.Register(protocol.CmdProfileInfo, func(ctx context.Context, args map[string]any, w io.Writer) error {
		format := agent.StringArg(args, "format")
		if format == "" {
			format = "json"
		}
		return execProfileInfo(ctx, agent.StringArg(args, "name"), format, w)
	})

	exec.Register(protocol.CmdProfileStart, func(ctx context.Context, args map[string]any, w io.Writer) error {
		return execProfileStart(ctx, agent.StringArg(args, "name"), agent.BoolArg(args, "dry_run"), w)
	})

	exec.Register(protocol.CmdProfileStop, func(ctx context.Context, args map[string]any, w io.Writer) error {
		return execProfileStop(ctx, agent.StringArg(args, "name"), w)
	})

	exec.Register(protocol.CmdProfileRm, func(ctx context.Context, args map[string]any, w io.Writer) error {
		// force=true: destructive actions from the panel require explicit intent in the request
		return execProfileRm(ctx, agent.StringArg(args, "name"), true, nil, w)
	})

	exec.Register(protocol.CmdMissionsPull, func(ctx context.Context, args map[string]any, w io.Writer) error {
		return execMissionsPull(ctx, agent.StringArg(args, "name"), agent.BoolArg(args, "dry_run"), w)
	})

	exec.Register(protocol.CmdUpdate, func(ctx context.Context, args map[string]any, w io.Writer) error {
		return execUpdate(ctx, version, w)
	})

	return exec
}

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.Flags().String("url", "", "WebSocket URL of the web panel (required)")
	agentCmd.Flags().String("token", "", "shared secret token (required)")
	agentCmd.Flags().String("name", "", "agent name shown in the panel (default: hostname)")
	agentCmd.Flags().Duration("heartbeat", 30*time.Second, "heartbeat interval")
	agentCmd.Flags().Duration("reconnect", 5*time.Second, "reconnect delay on disconnect")
}
