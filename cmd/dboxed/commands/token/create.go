package token

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Name string `help:"Specify the token name. Must be unique." required:""`

	ForWorkspace bool    `help:"If set, the token will be for the whole workspace" xor:"for"`
	Box          *string `help:"Specify box for which to create the token" xor:"for"`
	Machine      *string `help:"Specify machine for which to create the token" xor:"for"`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.TokenClient{Client: c}

	req := models.CreateToken{
		Name: cmd.Name,
	}

	if cmd.ForWorkspace {
		req.Type = dmodel.TokenTypeWorkspace
	} else if cmd.Box != nil {
		req.Type = dmodel.TokenTypeBox
		b, err := commandutils.GetBox(ctx, c, *cmd.Box)
		if err != nil {
			return err
		}
		req.BoxID = &b.ID
	} else if cmd.Machine != nil {
		req.Type = dmodel.TokenTypeMachine
		m, err := commandutils.GetMachine(ctx, c, *cmd.Machine)
		if err != nil {
			return err
		}
		req.MachineID = &m.ID
	} else {
		return fmt.Errorf("did not specify for what the token should be")
	}

	token, err := c2.CreateToken(ctx, req)
	if err != nil {
		return err
	}

	renderTokenCreated(token)

	return nil
}

func renderTokenCreated(token *models.Token) {
	// Define styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10")). // Green
		MarginTop(1).
		MarginBottom(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")). // Cyan
		Padding(1, 2).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // Gray
		Width(12).
		Align(lipgloss.Right)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")). // White
		Bold(true)

	tokenStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")). // Yellow
		Bold(true).
		Background(lipgloss.Color("235")). // Dark gray background
		Padding(0, 1)

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")). // Red
		Bold(true).
		MarginTop(1)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // Gray
		Italic(true)

	// Build the content
	var content strings.Builder

	// Token details
	content.WriteString(fmt.Sprintf("%s  %s\n",
		labelStyle.Render("Name:"),
		valueStyle.Render(token.Name),
	))

	content.WriteString(fmt.Sprintf("%s  %s\n",
		labelStyle.Render("ID:"),
		valueStyle.Render(token.ID),
	))

	content.WriteString(fmt.Sprintf("%s  %s\n",
		labelStyle.Render("Workspace:"),
		valueStyle.Render(token.Workspace),
	))

	scope := "Workspace"
	if token.MachineID != nil {
		scope = fmt.Sprintf("Machine %s", *token.MachineID)
	} else if token.BoxID != nil {
		scope = fmt.Sprintf("Box %s", *token.BoxID)
	} else if token.LoadBalancerId != nil {
		scope = fmt.Sprintf("Load Balancer %s", *token.LoadBalancerId)
	}
	content.WriteString(fmt.Sprintf("%s  %s\n",
		labelStyle.Render("Scope:"),
		valueStyle.Render(scope),
	))

	content.WriteString(fmt.Sprintf("%s  %s\n",
		labelStyle.Render("Created:"),
		valueStyle.Render(token.CreatedAt.String()),
	))

	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("%s  %s",
		labelStyle.Render("Token:"),
		tokenStyle.Render(*token.Token),
	))

	// Print the output
	fmt.Println(titleStyle.Render("✓ Token Created Successfully"))
	fmt.Println(boxStyle.Render(content.String()))
	fmt.Println(warningStyle.Render("⚠ IMPORTANT: Keep this token secret!"))
	fmt.Println(infoStyle.Render("This token value cannot be retrieved again. Store it securely."))
	fmt.Println()
}
