package commandutils

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmOptions contains options for customizing a confirmation dialog
type ConfirmOptions struct {
	// Title is the main question to ask
	Title string
	// Description provides additional context
	Description string
	// Affirmative is the text for the "yes" option (default: "Yes")
	Affirmative string
	// Negative is the text for the "no" option (default: "No")
	Negative string
	// WarningMessage is an optional warning shown before the prompt
	WarningMessage string
	// InfoMessage is an optional info message shown before the prompt
	InfoMessage string
}

// Confirm displays a beautiful confirmation dialog using huh and lipgloss
// Returns true if the user confirms, false if they cancel
func Confirm(opts ConfirmOptions) (bool, error) {
	// Set defaults
	if opts.Affirmative == "" {
		opts.Affirmative = "Yes"
	}
	if opts.Negative == "" {
		opts.Negative = "No"
	}
	if opts.Title == "" {
		opts.Title = "Are you sure?"
	}

	// Display warning message if provided
	if opts.WarningMessage != "" {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")). // Red
			Bold(true)

		fmt.Println()
		fmt.Println(warningStyle.Render(opts.WarningMessage))
	}

	// Display info message if provided
	if opts.InfoMessage != "" {
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")). // Gray
			Italic(true)

		fmt.Println(infoStyle.Render(opts.InfoMessage))
	}

	if opts.WarningMessage != "" || opts.InfoMessage != "" {
		fmt.Println()
	}

	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(opts.Title).
				Description(opts.Description).
				Affirmative(opts.Affirmative).
				Negative(opts.Negative).
				Value(&confirmed),
		),
	)

	err := form.Run()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}

// ConfirmDanger is a convenience function for dangerous operations
// It displays a red warning and uses strong affirmative language
func ConfirmDanger(title, warning string) (bool, error) {
	return Confirm(ConfirmOptions{
		Title:          title,
		Description:    "This action cannot be undone.",
		Affirmative:    "Yes, I'm sure",
		Negative:       "No, cancel",
		WarningMessage: "⚠ " + warning,
	})
}

// ConfirmSimple is a convenience function for simple yes/no questions
func ConfirmSimple(question string) (bool, error) {
	return Confirm(ConfirmOptions{
		Title:       question,
		Affirmative: "Yes",
		Negative:    "No",
	})
}
