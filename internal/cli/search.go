package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/itda-skills/jindo/internal/agent"
	"github.com/itda-skills/jindo/internal/command"
	"github.com/itda-skills/jindo/internal/skill"
	"github.com/spf13/cobra"
)

var (
	searchSkillsOnly   bool
	searchCommandsOnly bool
	searchAgentsOnly   bool
	searchNameOnly     bool
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search across skills, commands, and agents",
	Long: `Search for a keyword across all skills, commands, and agents.

Searches in name, description, and content by default.
Results are grouped by resource type.`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().BoolVarP(&searchSkillsOnly, "skills", "s", false, "Search only in skills")
	searchCmd.Flags().BoolVarP(&searchCommandsOnly, "commands", "c", false, "Search only in commands")
	searchCmd.Flags().BoolVarP(&searchAgentsOnly, "agents", "a", false, "Search only in agents")
	searchCmd.Flags().BoolVarP(&searchNameOnly, "name", "n", false, "Search only in names")
}

// SearchResult represents a single search result
type SearchResult struct {
	Type        string // "skill", "command", "agent"
	Name        string
	Description string
	Path        string
	MatchIn     string // where the match was found: "name", "description", "content"
}

func runSearch(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	query := strings.ToLower(args[0])

	var results []SearchResult

	// Determine which resources to search
	searchAll := !searchSkillsOnly && !searchCommandsOnly && !searchAgentsOnly

	// Search skills
	if searchAll || searchSkillsOnly {
		skillResults, err := searchSkills(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to search skills: %v\n", err)
		}
		results = append(results, skillResults...)
	}

	// Search commands
	if searchAll || searchCommandsOnly {
		cmdResults, err := searchCommands(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to search commands: %v\n", err)
		}
		results = append(results, cmdResults...)
	}

	// Search agents
	if searchAll || searchAgentsOnly {
		agentResults, err := searchAgents(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to search agents: %v\n", err)
		}
		results = append(results, agentResults...)
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	// Print results grouped by type
	printGroupedResults(results)

	return nil
}

func searchSkills(query string) ([]SearchResult, error) {
	store := skill.NewStore("~/.claude/skills")
	skills, err := store.List()
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, s := range skills {
		matchIn := matchSkill(s, query, store)
		if matchIn != "" {
			results = append(results, SearchResult{
				Type:        "skill",
				Name:        s.Name,
				Description: s.Description,
				Path:        s.Path,
				MatchIn:     matchIn,
			})
		}
	}

	return results, nil
}

func matchSkill(s *skill.Skill, query string, store *skill.Store) string {
	// Check name
	if strings.Contains(strings.ToLower(s.Name), query) {
		return "name"
	}

	if searchNameOnly {
		return ""
	}

	// Check description
	if strings.Contains(strings.ToLower(s.Description), query) {
		return "description"
	}

	// Check content
	content, err := store.GetContent(s.Name)
	if err == nil && strings.Contains(strings.ToLower(content), query) {
		return "content"
	}

	return ""
}

func searchCommands(query string) ([]SearchResult, error) {
	store := command.NewStore("~/.claude/commands")
	commands, err := store.List()
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, cmd := range commands {
		matchIn := matchCommand(cmd, query, store)
		if matchIn != "" {
			results = append(results, SearchResult{
				Type:        "command",
				Name:        cmd.Name,
				Description: cmd.Description,
				Path:        cmd.Path,
				MatchIn:     matchIn,
			})
		}
	}

	return results, nil
}

func matchCommand(cmd *command.Command, query string, store *command.Store) string {
	// Check name
	if strings.Contains(strings.ToLower(cmd.Name), query) {
		return "name"
	}

	if searchNameOnly {
		return ""
	}

	// Check description
	if strings.Contains(strings.ToLower(cmd.Description), query) {
		return "description"
	}

	// Check content
	content, err := store.GetContent(cmd.Name)
	if err == nil && strings.Contains(strings.ToLower(content), query) {
		return "content"
	}

	return ""
}

func searchAgents(query string) ([]SearchResult, error) {
	store := agent.NewStore("~/.claude/agents")
	agents, err := store.List()
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, a := range agents {
		matchIn := matchAgent(a, query, store)
		if matchIn != "" {
			results = append(results, SearchResult{
				Type:        "agent",
				Name:        a.Name,
				Description: a.Description,
				Path:        a.Path,
				MatchIn:     matchIn,
			})
		}
	}

	return results, nil
}

func matchAgent(a *agent.Agent, query string, store *agent.Store) string {
	// Check name
	if strings.Contains(strings.ToLower(a.Name), query) {
		return "name"
	}

	if searchNameOnly {
		return ""
	}

	// Check description
	if strings.Contains(strings.ToLower(a.Description), query) {
		return "description"
	}

	// Check content
	content, err := store.GetContent(a.Name)
	if err == nil && strings.Contains(strings.ToLower(content), query) {
		return "content"
	}

	return ""
}

func printGroupedResults(results []SearchResult) {
	// Group by type
	skillResults := filterByType(results, "skill")
	cmdResults := filterByType(results, "command")
	agentResults := filterByType(results, "agent")

	total := len(results)

	if len(skillResults) > 0 {
		fmt.Printf("Skills (%d):\n", len(skillResults))
		for _, r := range skillResults {
			printResult(r)
		}
		fmt.Println()
	}

	if len(cmdResults) > 0 {
		fmt.Printf("Commands (%d):\n", len(cmdResults))
		for _, r := range cmdResults {
			printResult(r)
		}
		fmt.Println()
	}

	if len(agentResults) > 0 {
		fmt.Printf("Agents (%d):\n", len(agentResults))
		for _, r := range agentResults {
			printResult(r)
		}
		fmt.Println()
	}

	fmt.Printf("Total: %d results\n", total)
}

func filterByType(results []SearchResult, typ string) []SearchResult {
	var filtered []SearchResult
	for _, r := range results {
		if r.Type == typ {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func printResult(r SearchResult) {
	desc := r.Description
	if len(desc) > 50 {
		desc = desc[:47] + "..."
	}
	fmt.Printf("  %-20s  %s  (match in %s)\n", r.Name, desc, r.MatchIn)
}
