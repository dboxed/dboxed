package box

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type StatusCmd struct {
	Box string `help:"Specify the box" required:"" arg:""`
}

type PrintDockerContainer struct {
	ContainerID string `col:"Container ID"`
	Names       string `col:"Names"`
	Image       string `col:"Image"`
	State       string `col:"State"`
	Status      string `col:"Status"`
	Ports       string `col:"Ports"`
	CreatedAt   string `col:"Created"`
}

func (cmd *StatusCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}
	runStatus, err := c2.GetBoxRunStatus(ctx, b.ID)
	if err != nil {
		return err
	}

	// Display box status with styled output
	renderBoxStatus(b, runStatus)

	// Display docker containers table
	if runStatus.DockerPs != nil && len(runStatus.DockerPs) > 0 {
		containers, err := parseDockerPs(runStatus.DockerPs)
		if err != nil {
			return err
		}

		if len(containers) > 0 {
			err = commandutils.PrintTable(os.Stdout, containers)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func renderBoxStatus(box *models.Box, runStatus *models.BoxRunStatus) {
	// Define styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(16).
		Align(lipgloss.Right)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	statusColors := map[string]lipgloss.Color{
		"running": lipgloss.Color("10"),
		"stopped": lipgloss.Color("9"),
		"paused":  lipgloss.Color("11"),
		"up":      lipgloss.Color("10"),
		"down":    lipgloss.Color("9"),
	}

	// Title
	fmt.Println(titleStyle.Render(fmt.Sprintf("Box Status: %s", box.Name)))

	// Box ID
	fmt.Printf("%s  %s\n",
		labelStyle.Render("Box ID:"),
		valueStyle.Render(fmt.Sprintf("%d", box.ID)),
	)

	// Desired State with color
	desiredStateColor := lipgloss.Color("241") // default gray
	if color, ok := statusColors[box.DesiredState]; ok {
		desiredStateColor = color
	}
	desiredStateStyle := valueStyle.Copy().Foreground(desiredStateColor)
	fmt.Printf("%s  %s\n",
		labelStyle.Render("Desired State:"),
		desiredStateStyle.Render(box.DesiredState),
	)

	// Run Status with color
	runStatusValue := formatOptionalString(runStatus.RunStatus)
	statusColor := lipgloss.Color("241") // default gray
	if color, ok := statusColors[runStatusValue]; ok {
		statusColor = color
	}
	statusValueStyle := valueStyle.Copy().Foreground(statusColor)
	fmt.Printf("%s  %s\n",
		labelStyle.Render("Run Status:"),
		statusValueStyle.Render(runStatusValue),
	)

	// Start Time
	fmt.Printf("%s  %s\n",
		labelStyle.Render("Start Time:"),
		valueStyle.Render(formatOptionalTime(runStatus.StartTime)),
	)

	// Stop Time
	fmt.Printf("%s  %s\n",
		labelStyle.Render("Stop Time:"),
		valueStyle.Render(formatOptionalTime(runStatus.StopTime)),
	)

	// Status Time
	fmt.Printf("%s  %s\n",
		labelStyle.Render("Status Time:"),
		valueStyle.Render(formatOptionalTime(runStatus.StatusTime)),
	)

	fmt.Println() // Empty line before containers table
}

func formatOptionalString(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}

func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.String()
}

func colorizeState(state string) string {
	style := lipgloss.NewStyle()
	switch state {
	case "running":
		style = style.Foreground(lipgloss.Color("10")) // Green
	case "exited":
		style = style.Foreground(lipgloss.Color("11")) // Yellow
	default:
		style = style.Foreground(lipgloss.Color("15")) // White
	}
	return style.Render(state)
}

func parseDockerPs(compressed []byte) ([]PrintDockerContainer, error) {
	// Decompress the gzip data
	r, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Parse newline-delimited JSON
	var containers []PrintDockerContainer
	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		var container dockercli.DockerPS
		err := decoder.Decode(&container)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Colorize state based on value
		stateStyled := colorizeState(container.State)

		containers = append(containers, PrintDockerContainer{
			ContainerID: container.ID,
			Names:       container.Names,
			Image:       container.Image,
			State:       stateStyled,
			Status:      container.Status,
			Ports:       container.Ports,
			CreatedAt:   container.CreatedAt,
		})
	}

	return containers, nil
}
