package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itda-work/itda-jindo/internal/pkg/pkgmgr"
	"github.com/itda-work/itda-jindo/internal/pkg/repo"
)

// Tab represents a tab in the TUI
type Tab int

const (
	TabSkills Tab = iota
	TabCommands
	TabAgents
	TabHooks
)

func (t Tab) String() string {
	switch t {
	case TabSkills:
		return "Skills"
	case TabCommands:
		return "Commands"
	case TabAgents:
		return "Agents"
	case TabHooks:
		return "Hooks"
	default:
		return ""
	}
}

func (t Tab) PackageType() repo.PackageType {
	switch t {
	case TabSkills:
		return repo.TypeSkill
	case TabCommands:
		return repo.TypeCommand
	case TabAgents:
		return repo.TypeAgent
	case TabHooks:
		return repo.TypeHook
	default:
		return ""
	}
}

// PackageItem represents a package in the list
type PackageItem struct {
	Namespace   string
	Name        string
	Path        string
	Type        repo.PackageType
	IsInstalled bool
	HasUpdate   bool
	Selected    bool
}

// Model represents the TUI state
type Model struct {
	tabs      []Tab
	activeTab Tab
	items     map[Tab][]PackageItem
	cursor    int
	width     int
	height    int
	manager   *pkgmgr.Manager
	message   string
	quitting  bool
}

// Styles
var (
	tabStyle = lipgloss.NewStyle().
			Padding(0, 2)

	activeTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Bold(true).
			Foreground(lipgloss.Color("205"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99"))

	namespaceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	installedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	updateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)
)

// Key bindings
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Tab      key.Binding
	Select   key.Binding
	Install  key.Binding
	SelectAll key.Binding
	Quit     key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "prev tab"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next tab"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	Select: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "select"),
	),
	Install: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "install"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "select all"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q/esc", "quit"),
	),
}

// NewModel creates a new browse TUI model
func NewModel(manager *pkgmgr.Manager) *Model {
	return &Model{
		tabs:      []Tab{TabSkills, TabCommands, TabAgents, TabHooks},
		activeTab: TabSkills,
		items:     make(map[Tab][]PackageItem),
		manager:   manager,
	}
}

// LoadPackages loads packages from all registered repositories
func (m *Model) LoadPackages() error {
	repoStore := m.manager.RepoStore()
	repos, err := repoStore.List()
	if err != nil {
		return err
	}

	// Get installed packages
	installed, err := m.manager.List()
	if err != nil {
		return err
	}
	installedMap := make(map[string]bool)
	for _, pkg := range installed {
		installedMap[pkg.Name] = true
	}

	// Initialize items map
	for _, tab := range m.tabs {
		m.items[tab] = []PackageItem{}
	}

	// Load packages from each repository
	for _, r := range repos {
		items, err := repoStore.Browse(r.Namespace, "")
		if err != nil {
			continue
		}

		for _, item := range items {
			var tab Tab
			switch item.Type {
			case repo.TypeSkill:
				tab = TabSkills
			case repo.TypeCommand:
				tab = TabCommands
			case repo.TypeAgent:
				tab = TabAgents
			case repo.TypeHook:
				tab = TabHooks
			default:
				continue
			}

			namespacedName := pkgmgr.MakeNamespacedName(r.Namespace, item.Name)
			pkgItem := PackageItem{
				Namespace:   r.Namespace,
				Name:        item.Name,
				Path:        item.Path,
				Type:        item.Type,
				IsInstalled: installedMap[namespacedName],
			}
			m.items[tab] = append(m.items[tab], pkgItem)
		}
	}

	return nil
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Clear message on any key press
		m.message = ""

		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, keys.Tab), key.Matches(msg, keys.Right):
			m.activeTab = Tab((int(m.activeTab) + 1) % len(m.tabs))
			m.cursor = 0
			return m, nil

		case key.Matches(msg, keys.Left):
			m.activeTab = Tab((int(m.activeTab) - 1 + len(m.tabs)) % len(m.tabs))
			m.cursor = 0
			return m, nil

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case key.Matches(msg, keys.Down):
			items := m.items[m.activeTab]
			if m.cursor < len(items)-1 {
				m.cursor++
			}
			return m, nil

		case key.Matches(msg, keys.Select):
			items := m.items[m.activeTab]
			if m.cursor < len(items) {
				item := &m.items[m.activeTab][m.cursor]
				if !item.IsInstalled {
					item.Selected = !item.Selected
				}
			}
			return m, nil

		case key.Matches(msg, keys.SelectAll):
			items := m.items[m.activeTab]
			// Check if all non-installed items are selected
			allSelected := true
			for _, item := range items {
				if !item.IsInstalled && !item.Selected {
					allSelected = false
					break
				}
			}
			// Toggle all
			for i := range m.items[m.activeTab] {
				if !m.items[m.activeTab][i].IsInstalled {
					m.items[m.activeTab][i].Selected = !allSelected
				}
			}
			return m, nil

		case key.Matches(msg, keys.Install):
			return m, m.installSelected()
		}
	}

	return m, nil
}

// installSelected installs selected packages
func (m *Model) installSelected() tea.Cmd {
	return func() tea.Msg {
		var installedCount int
		for tab := range m.items {
			for i := range m.items[tab] {
				item := &m.items[tab][i]
				if item.Selected && !item.IsInstalled {
					spec := fmt.Sprintf("%s:%s", item.Namespace, item.Path)
					_, err := m.manager.Install(spec)
					if err == nil {
						item.IsInstalled = true
						item.Selected = false
						installedCount++
					}
				}
			}
		}
		if installedCount > 0 {
			m.message = fmt.Sprintf("Installed %d package(s)", installedCount)
		}
		return nil
	}
}

// View renders the UI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("jd pkg browse"))
	b.WriteString("\n\n")

	// Tabs
	var tabs []string
	for _, tab := range m.tabs {
		style := tabStyle
		if tab == m.activeTab {
			style = activeTabStyle
		}
		tabs = append(tabs, style.Render(fmt.Sprintf("[%s]", tab.String())))
	}
	b.WriteString(strings.Join(tabs, " "))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n\n")

	// Items
	items := m.items[m.activeTab]
	if len(items) == 0 {
		b.WriteString(helpStyle.Render("  No packages found"))
		b.WriteString("\n")
	} else {
		// Group by namespace
		namespaces := make(map[string][]int)
		var nsOrder []string
		for i, item := range items {
			if _, exists := namespaces[item.Namespace]; !exists {
				nsOrder = append(nsOrder, item.Namespace)
			}
			namespaces[item.Namespace] = append(namespaces[item.Namespace], i)
		}

		globalIdx := 0
		for _, ns := range nsOrder {
			b.WriteString(namespaceStyle.Render(ns))
			b.WriteString("\n")

			for _, idx := range namespaces[ns] {
				item := items[idx]
				cursor := "  "
				if globalIdx == m.cursor {
					cursor = "> "
				}

				checkbox := "[ ]"
				if item.IsInstalled {
					checkbox = "[✓]"
				} else if item.Selected {
					checkbox = "[*]"
				}

				status := ""
				if item.IsInstalled {
					status = installedStyle.Render("installed")
				} else if item.HasUpdate {
					status = updateStyle.Render("update available")
				}

				name := item.Name
				if globalIdx == m.cursor {
					name = selectedStyle.Render(name)
				}

				line := fmt.Sprintf("%s %s %s", cursor, checkbox, name)
				if status != "" {
					padding := m.width - lipgloss.Width(line) - lipgloss.Width(status) - 2
					if padding < 1 {
						padding = 1
					}
					line += strings.Repeat(" ", padding) + status
				}
				b.WriteString(line)
				b.WriteString("\n")
				globalIdx++
			}
			b.WriteString("\n")
		}
	}

	// Message
	if m.message != "" {
		b.WriteString("\n")
		b.WriteString(messageStyle.Render(m.message))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")
	help := helpStyle.Render("↑/↓: navigate  ←/→/tab: switch tab  space: select  a: select all  enter: install  q: quit")
	b.WriteString(help)
	b.WriteString("\n")

	return b.String()
}

// Run starts the TUI
func Run(manager *pkgmgr.Manager) error {
	m := NewModel(manager)
	if err := m.LoadPackages(); err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
