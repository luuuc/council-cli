package mcp

import (
	"context"
	"fmt"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/expert"
	"github.com/luuuc/council-cli/internal/export"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ExpertURIPrefix is the URI scheme prefix for expert resources
const ExpertURIPrefix = "council://experts/"

// Server wraps the MCP server with council-specific functionality
type Server struct {
	mcp *server.MCPServer
}

// NewServer creates a new MCP server configured for council
func NewServer() *Server {
	s := server.NewMCPServer(
		"council",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
		server.WithPromptCapabilities(true),
	)

	srv := &Server{mcp: s}
	srv.registerTools()
	srv.registerResources()
	srv.registerPrompts()

	return srv
}

// ServeStdio starts the server using stdio transport
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcp)
}

func (s *Server) registerTools() {
	// list_experts tool
	listExpertsTool := mcp.NewTool("list_experts",
		mcp.WithDescription("List all experts in the council"),
	)
	s.mcp.AddTool(listExpertsTool, s.handleListExperts)

	// get_expert tool
	getExpertTool := mcp.NewTool("get_expert",
		mcp.WithDescription("Get details of a specific expert"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The expert ID (e.g., 'dhh', 'kent-beck')"),
		),
	)
	s.mcp.AddTool(getExpertTool, s.handleGetExpert)

	// consult_council tool
	consultTool := mcp.NewTool("consult_council",
		mcp.WithDescription("Get perspectives from all council experts on a topic"),
		mcp.WithString("topic",
			mcp.Description("Optional topic to focus the consultation on"),
		),
	)
	s.mcp.AddTool(consultTool, s.handleConsultCouncil)
}

func (s *Server) registerResources() {
	// Dynamic resource template for individual experts
	template := mcp.NewResourceTemplate(
		ExpertURIPrefix+"{id}",
		"Expert Profile",
		mcp.WithTemplateMIMEType("text/markdown"),
		mcp.WithTemplateDescription("Individual expert persona from the council"),
	)
	s.mcp.AddResourceTemplate(template, s.handleExpertResource)
}

func (s *Server) registerPrompts() {
	// /council prompt
	councilPrompt := mcp.NewPrompt("council",
		mcp.WithPromptDescription("Review work with your expert council"),
		mcp.WithArgument("content",
			mcp.ArgumentDescription("The content or code to review"),
			mcp.RequiredArgument(),
		),
	)
	s.mcp.AddPrompt(councilPrompt, s.handleCouncilPrompt)
}

func (s *Server) handleListExperts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !config.Exists() {
		return mcp.NewToolResultError("council not initialized: run 'council init' first"), nil
	}

	experts, err := expert.List()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list experts: %v", err)), nil
	}

	if len(experts) == 0 {
		return mcp.NewToolResultText("No experts in council. Run 'council setup --apply' to add experts."), nil
	}

	var result string
	for _, e := range experts {
		result += fmt.Sprintf("- **%s** (%s): %s\n", e.Name, e.ID, e.Focus)
	}

	return mcp.NewToolResultText(result), nil
}

func (s *Server) handleGetExpert(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !config.Exists() {
		return mcp.NewToolResultError("council not initialized: run 'council init' first"), nil
	}

	id, err := request.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: id"), nil
	}

	e, err := expert.Load(id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("expert '%s' not found", id)), nil
	}

	result := fmt.Sprintf("# %s\n\n**Focus**: %s\n\n", e.Name, e.Focus)

	if e.Philosophy != "" {
		result += fmt.Sprintf("## Philosophy\n\n%s\n\n", e.Philosophy)
	}

	if len(e.Principles) > 0 {
		result += "## Principles\n\n"
		for _, p := range e.Principles {
			result += fmt.Sprintf("- %s\n", p)
		}
		result += "\n"
	}

	if len(e.RedFlags) > 0 {
		result += "## Red Flags\n\n"
		for _, r := range e.RedFlags {
			result += fmt.Sprintf("- %s\n", r)
		}
		result += "\n"
	}

	return mcp.NewToolResultText(result), nil
}

func (s *Server) handleConsultCouncil(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !config.Exists() {
		return mcp.NewToolResultError("council not initialized: run 'council init' first"), nil
	}

	experts, err := expert.List()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list experts: %v", err)), nil
	}

	if len(experts) == 0 {
		return mcp.NewToolResultText("No experts in council. Run 'council setup --apply' to add experts."), nil
	}

	// Use existing export format
	result := export.FormatMarkdown(experts)

	return mcp.NewToolResultText(result), nil
}

func (s *Server) handleExpertResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Extract ID from URI (council://experts/{id})
	uri := request.Params.URI
	id := extractExpertID(uri)

	if id == "" {
		return nil, fmt.Errorf("invalid expert URI: %s", uri)
	}

	e, err := expert.Load(id)
	if err != nil {
		return nil, fmt.Errorf("expert '%s' not found", id)
	}

	content := fmt.Sprintf("# %s\n\n**Focus**: %s\n\n%s", e.Name, e.Focus, e.Body)

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     content,
		},
	}, nil
}

func (s *Server) handleCouncilPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	content := request.Params.Arguments["content"]
	if content == "" {
		content = "[Please provide content to review]"
	}

	experts, err := expert.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list experts: %v", err)
	}

	// Build expert context
	expertContext := export.FormatMarkdown(experts)

	messages := []mcp.PromptMessage{
		mcp.NewPromptMessage(
			mcp.RoleUser,
			mcp.NewTextContent(fmt.Sprintf(`You have access to an expert council. When reviewing work, consider each expert's perspective and provide feedback as if channeling their expertise.

%s

---

Please review the following content from each expert's perspective:

%s`, expertContext, content)),
		),
	}

	return mcp.NewGetPromptResult(
		"Expert Council Review",
		messages,
	), nil
}

// extractExpertID extracts the expert ID from a council://experts/{id} URI
func extractExpertID(uri string) string {
	if len(uri) > len(ExpertURIPrefix) {
		return uri[len(ExpertURIPrefix):]
	}
	return ""
}
