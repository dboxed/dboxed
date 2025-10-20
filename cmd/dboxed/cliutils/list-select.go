package cliutils

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxListHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type Item interface {
	list.Item

	GetId() int64
	GetName() string
}

type itemDelegate[T Item] struct{}

func (d itemDelegate[T]) Height() int                             { return 1 }
func (d itemDelegate[T]) Spacing() int                            { return 0 }
func (d itemDelegate[T]) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate[T]) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.GetName())

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model[T Item] struct {
	list         list.Model
	selectedItem *T
	quitting     bool
}

func (m model[T]) Init() tea.Cmd {
	return nil
}

func (m model[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(T)
			if ok {
				m.selectedItem = &i
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model[T]) View() string {
	if m.selectedItem != nil {
		return quitTextStyle.Render(fmt.Sprintf("Selected %s.", (*m.selectedItem).GetName()))
	}
	if m.quitting {
		return quitTextStyle.Render("Aborted.")
	}
	return "\n" + m.list.View()
}

func ListSelect[T Item](title string, items []list.Item) (T, error) {
	const defaultWidth = 20
	listHeight := min(8+len(items), maxListHeight)

	l := list.New(items, itemDelegate[T]{}, defaultWidth, listHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model[T]{list: l}

	var zero T
	newModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return zero, err
	}
	nm, ok := newModel.(model[T])
	if !ok {
		return zero, fmt.Errorf("unexpected model returned")
	}
	if nm.selectedItem == nil {
		return zero, fmt.Errorf("aborted selection")
	}

	return *nm.selectedItem, nil
}
