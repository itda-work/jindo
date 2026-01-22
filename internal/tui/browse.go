package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itda-work/jindo/internal/pkg/pkgmgr"
	"github.com/itda-work/jindo/internal/pkg/repo"
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
	LocalPath   string // Full local path for preview
	Type        repo.PackageType
	IsInstalled bool
	HasUpdate   bool
	Selected    bool
}

// installDoneMsg is sent when installation completes
type installDoneMsg struct {
	count  int
	errors []string
}

// uninstallDoneMsg is sent when uninstallation completes
type uninstallDoneMsg struct {
	success bool
	name    string
	err     error
}

// Model represents the TUI state
type Model struct {
	tabs                []Tab
	activeTab           Tab
	items               map[Tab][]PackageItem
	cursor              int
	listOffset          int    // Scroll offset for list panel
	width               int
	height              int
	manager             *pkgmgr.Manager
	message             string
	quitting            bool
	preview             string // Cached preview content
	namespaceFilter     string // Filter by namespace (empty = all)
	installing          bool   // True while installation is in progress
	confirmingUninstall bool   // True when waiting for uninstall confirmation
	confirmingItem      *PackageItem
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

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	previewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				MarginBottom(1)

	previewBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("241")).
				Padding(1)

	previewContentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	listPaneStyle = lipgloss.NewStyle().
			Padding(0, 1)
)

// Key bindings
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Tab       key.Binding
	Select    key.Binding
	Install   key.Binding
	Uninstall key.Binding
	SelectAll key.Binding
	Quit      key.Binding
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
	Uninstall: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "uninstall"),
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
		// Apply namespace filter
		if m.namespaceFilter != "" && r.Namespace != m.namespaceFilter {
			continue
		}

		items, err := repoStore.Browse(r.Namespace, "")
		if err != nil {
			continue
		}

		// Get local repo path for preview
		repoLocalPath, err := repoStore.RepoLocalPath(r.Namespace)
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

			// Determine the file to preview
			localPath := filepath.Join(repoLocalPath, item.Path)
			if item.Type == repo.TypeSkill {
				// Skills are directories, look for SKILL.md
				for _, name := range []string{"SKILL.md", "skill.md"} {
					candidate := filepath.Join(localPath, name)
					if _, err := os.Stat(candidate); err == nil {
						localPath = candidate
						break
					}
				}
			}

			pkgItem := PackageItem{
				Namespace:   r.Namespace,
				Name:        item.Name,
				Path:        item.Path,
				LocalPath:   localPath,
				Type:        item.Type,
				IsInstalled: installedMap[namespacedName],
			}
			m.items[tab] = append(m.items[tab], pkgItem)
		}
	}

	// Load initial preview
	m.updatePreview()

	return nil
}

// loadPreview loads and returns preview content from a file (max 30 lines)
func loadPreview(path string, maxLines int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("Unable to load preview:\n%v", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "\n... (truncated)")
	}

	return strings.Join(lines, "\n")
}

// updatePreview updates the preview content for the current selection
func (m *Model) updatePreview() {
	items := m.items[m.activeTab]
	if len(items) == 0 || m.cursor >= len(items) {
		m.preview = "No package selected"
		return
	}

	item := items[m.cursor]
	m.preview = loadPreview(item.LocalPath, 50)
}

// listVisibleHeight returns the number of visible lines in the list panel
func (m *Model) listVisibleHeight() int {
	headerHeight := 5  // title + tabs + separator
	footerHeight := 4  // message + help
	contentHeight := m.height - headerHeight - footerHeight
	if contentHeight < 5 {
		contentHeight = 5
	}
	return contentHeight
}

// adjustListScroll adjusts listOffset to keep cursor visible
func (m *Model) adjustListScroll() {
	visibleHeight := m.listVisibleHeight()

	// Count total lines including namespace headers
	items := m.items[m.activeTab]
	namespaces := make(map[string]bool)
	lineIndex := 0
	cursorLine := 0

	for i, item := range items {
		if !namespaces[item.Namespace] {
			namespaces[item.Namespace] = true
			lineIndex++ // namespace header line
		}
		if i == m.cursor {
			cursorLine = lineIndex
		}
		lineIndex++
	}

	// Adjust offset to keep cursor in view
	if cursorLine < m.listOffset {
		m.listOffset = cursorLine
	}
	if cursorLine >= m.listOffset+visibleHeight {
		m.listOffset = cursorLine - visibleHeight + 1
	}
	if m.listOffset < 0 {
		m.listOffset = 0
	}
}

// getCurrentItem returns the currently selected item or nil
func (m *Model) getCurrentItem() *PackageItem {
	items := m.items[m.activeTab]
	if len(items) == 0 || m.cursor >= len(items) {
		return nil
	}
	return &items[m.cursor]
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

	case installDoneMsg:
		m.installing = false
		if msg.count > 0 {
			m.message = fmt.Sprintf("✓ Installed %d package(s)", msg.count)
		}
		if len(msg.errors) > 0 {
			if m.message != "" {
				m.message += " | "
			}
			m.message += fmt.Sprintf("✗ %d error(s)", len(msg.errors))
		}
		return m, nil

	case uninstallDoneMsg:
		m.confirmingUninstall = false
		if msg.success {
			// Update item state in all tabs
			for tab := range m.items {
				for i := range m.items[tab] {
					item := &m.items[tab][i]
					namespacedName := pkgmgr.MakeNamespacedName(item.Namespace, item.Name)
					if namespacedName == msg.name {
						item.IsInstalled = false
						item.Selected = false
						break
					}
				}
			}
			m.message = fmt.Sprintf("✓ Uninstalled %s", msg.name)
		} else {
			m.message = fmt.Sprintf("✗ Failed to uninstall: %v", msg.err)
		}
		m.confirmingItem = nil
		return m, nil

	case tea.KeyMsg:
		// Ignore key presses while installing
		if m.installing {
			return m, nil
		}

		// Handle confirmation prompt
		if m.confirmingUninstall {
			switch msg.String() {
			case "y", "Y":
				// Proceed with uninstall
				m.message = "Uninstalling..."
				return m, m.uninstallPackage(m.confirmingItem)
			case "n", "N", "esc", "q":
				// Cancel
				m.confirmingUninstall = false
				m.confirmingItem = nil
				m.message = "Cancelled"
				return m, nil
			}
			// Ignore other keys during confirmation
			return m, nil
		}

		// Clear message on any key press
		m.message = ""

		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, keys.Tab), key.Matches(msg, keys.Right):
			m.activeTab = Tab((int(m.activeTab) + 1) % len(m.tabs))
			m.cursor = 0
			m.listOffset = 0
			m.updatePreview()
			return m, nil

		case key.Matches(msg, keys.Left):
			m.activeTab = Tab((int(m.activeTab) - 1 + len(m.tabs)) % len(m.tabs))
			m.cursor = 0
			m.listOffset = 0
			m.updatePreview()
			return m, nil

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
				m.adjustListScroll()
				m.updatePreview()
			}
			return m, nil

		case key.Matches(msg, keys.Down):
			items := m.items[m.activeTab]
			if m.cursor < len(items)-1 {
				m.cursor++
				m.adjustListScroll()
				m.updatePreview()
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
			// Check if any packages are selected
			hasSelected := false
			for tab := range m.items {
				for _, item := range m.items[tab] {
					if item.Selected && !item.IsInstalled {
						hasSelected = true
						break
					}
				}
				if hasSelected {
					break
				}
			}
			if !hasSelected {
				m.message = "No packages selected"
				return m, nil
			}
			m.installing = true
			m.message = "Installing..."
			return m, m.installSelected()

		case key.Matches(msg, keys.Uninstall):
			item := m.getCurrentItem()
			if item != nil && item.IsInstalled {
				// Show confirmation prompt
				m.confirmingUninstall = true
				m.confirmingItem = item
				namespacedName := pkgmgr.MakeNamespacedName(item.Namespace, item.Name)
				m.message = fmt.Sprintf("Uninstall '%s'? [y/N]", namespacedName)
				return m, nil
			}
			return m, nil
		}
	}

	return m, nil
}

// installSelected installs selected packages
func (m *Model) installSelected() tea.Cmd {
	return func() tea.Msg {
		var installedCount int
		var errors []string
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
					} else {
						errors = append(errors, fmt.Sprintf("%s: %v", item.Name, err))
					}
				}
			}
		}
		return installDoneMsg{count: installedCount, errors: errors}
	}
}

// uninstallPackage uninstalls a single package
func (m *Model) uninstallPackage(item *PackageItem) tea.Cmd {
	return func() tea.Msg {
		namespacedName := pkgmgr.MakeNamespacedName(item.Namespace, item.Name)
		err := m.manager.Uninstall(namespacedName)
		if err != nil {
			return uninstallDoneMsg{
				success: false,
				name:    namespacedName,
				err:     err,
			}
		}

		return uninstallDoneMsg{
			success: true,
			name:    namespacedName,
			err:     nil,
		}
	}
}

// renderList renders the left pane with package list
func (m Model) renderList(width, height int) string {
	var lines []string

	items := m.items[m.activeTab]
	if len(items) == 0 {
		lines = append(lines, helpStyle.Render("No packages found"))
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
			lines = append(lines, namespaceStyle.Render(ns))

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

				name := item.Name

				// Truncate name if too long (before applying styles)
				maxNameLen := width - 10
				if maxNameLen < 10 {
					maxNameLen = 10
				}
				if len(name) > maxNameLen {
					name = name[:maxNameLen-3] + "..."
				}

				// Apply style after truncation to preserve ANSI escape codes
				if globalIdx == m.cursor {
					name = selectedStyle.Render(name)
				}

				line := fmt.Sprintf("%s%s %s", cursor, checkbox, name)

				// Add status indicator
				if item.IsInstalled {
					line += " " + installedStyle.Render("✓")
				}

				lines = append(lines, line)
				globalIdx++
			}
		}
	}

	// Apply scroll offset
	totalLines := len(lines)
	startIdx := m.listOffset
	if startIdx > totalLines {
		startIdx = totalLines
	}
	endIdx := startIdx + height
	if endIdx > totalLines {
		endIdx = totalLines
	}

	visibleLines := lines[startIdx:endIdx]

	// Add scroll indicator if needed
	var result strings.Builder
	if startIdx > 0 {
		result.WriteString(helpStyle.Render("  ↑ more above"))
		result.WriteString("\n")
		if len(visibleLines) > 1 {
			visibleLines = visibleLines[1:]
		}
	}

	for _, line := range visibleLines {
		result.WriteString(line)
		result.WriteString("\n")
	}

	if endIdx < totalLines {
		result.WriteString(helpStyle.Render("  ↓ more below"))
		result.WriteString("\n")
	}

	return result.String()
}

// renderPreview renders the right pane with preview content
func (m Model) renderPreview(width, height int) string {
	var b strings.Builder

	item := m.getCurrentItem()
	if item == nil {
		return previewBorderStyle.Width(width - 4).Height(height - 4).Render("No package selected")
	}

	// Title
	title := previewTitleStyle.Render(fmt.Sprintf("📄 %s", item.Name))
	b.WriteString(title)
	b.WriteString("\n")

	// Path info
	pathInfo := helpStyle.Render(fmt.Sprintf("Path: %s", item.Path))
	b.WriteString(pathInfo)
	b.WriteString("\n\n")

	// Content
	content := m.preview
	// Wrap content to fit width
	contentLines := strings.Split(content, "\n")
	maxContentWidth := width - 6
	if maxContentWidth < 20 {
		maxContentWidth = 20
	}

	var wrappedLines []string
	for _, line := range contentLines {
		if len(line) > maxContentWidth {
			line = line[:maxContentWidth-3] + "..."
		}
		wrappedLines = append(wrappedLines, line)
	}

	// Limit height
	maxLines := height - 8
	if maxLines < 5 {
		maxLines = 5
	}
	if len(wrappedLines) > maxLines {
		wrappedLines = wrappedLines[:maxLines]
		wrappedLines = append(wrappedLines, "...")
	}

	b.WriteString(previewContentStyle.Render(strings.Join(wrappedLines, "\n")))

	return previewBorderStyle.Width(width - 4).Render(b.String())
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

	// Calculate pane dimensions (30:70 split)
	headerHeight := 5 // title + tabs + separator
	footerHeight := 4 // message + help
	contentHeight := m.height - headerHeight - footerHeight
	if contentHeight < 10 {
		contentHeight = 10
	}

	listWidth := m.width * 30 / 100
	if listWidth < 20 {
		listWidth = 20
	}
	previewWidth := m.width - listWidth - 2 // 2 for separator
	if previewWidth < 30 {
		previewWidth = 30
	}

	// Render list and preview panes
	listContent := m.renderList(listWidth, contentHeight)
	previewContent := m.renderPreview(previewWidth, contentHeight)

	// Style the list pane
	listPane := listPaneStyle.Width(listWidth).Height(contentHeight).Render(listContent)

	// Join panes horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, listPane, previewContent)
	b.WriteString(mainContent)
	b.WriteString("\n")

	// Message
	if m.message != "" {
		b.WriteString(messageStyle.Render(m.message))
		b.WriteString("\n")
	}

	// Help
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")
	help := helpStyle.Render("↑/↓: navigate  ←/→/tab: switch tab  space: select  a: select all  enter: install  d: uninstall  q: quit")
	b.WriteString(help)

	return b.String()
}

// Run starts the TUI
func Run(manager *pkgmgr.Manager, namespace string, startTab Tab) error {
	m := NewModel(manager)
	m.namespaceFilter = namespace
	m.activeTab = startTab
	if err := m.LoadPackages(); err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
