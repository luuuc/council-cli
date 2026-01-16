package cmd

import (
	"github.com/luuuc/council-cli/internal/mcp"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(mcpCmd)
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for Claude Desktop",
	Long: `Starts a local MCP server that Claude Desktop can connect to.

Configure in ~/Library/Application Support/Claude/claude_desktop_config.json:

{
  "mcpServers": {
    "council": {
      "command": "council",
      "args": ["mcp"]
    }
  }
}

The server exposes your local council via the MCP protocol:
- Resources: Each expert as council://experts/{id}
- Tools: list_experts, get_expert, consult_council
- Prompts: /council for expert review`,
	RunE: func(cmd *cobra.Command, args []string) error {
		server := mcp.NewServer()
		return server.ServeStdio()
	},
}
