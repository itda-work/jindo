package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/itda-work/itda-jindo/internal/pkg/repo"
	"github.com/spf13/cobra"
)

var pkgRepoAddNamespace string

var pkgRepoAddCmd = &cobra.Command{
	Use:   "add <gh:owner/repo>",
	Short: "Register a GitHub repository",
	Long: `Register a GitHub repository containing Claude Code packages.

The repository URL must be in the format: gh:owner/repo

A namespace will be automatically generated from the owner and repo names
(first 4 characters of each, joined by a hyphen). You can override this
with the --namespace flag.

Examples:
  jd pkg repo add gh:affaan-m/everything-claude-code
  jd pkg repo add gh:user/claude-skills --namespace mysk`,
	Args: cobra.ExactArgs(1),
	RunE: runPkgRepoAdd,
}

func init() {
	pkgRepoCmd.AddCommand(pkgRepoAddCmd)
	pkgRepoAddCmd.Flags().StringVarP(&pkgRepoAddNamespace, "namespace", "n", "", "Custom namespace for the repository")
}

func runPkgRepoAdd(_ *cobra.Command, args []string) error {
	url := args[0]

	// Parse URL to generate namespace if not provided
	owner, repoName, err := repo.ParseURL(url)
	if err != nil {
		return fmt.Errorf("invalid URL format. Use: gh:owner/repo")
	}

	store := repo.NewStore("~/.itda-jindo")

	namespace := pkgRepoAddNamespace
	if namespace == "" {
		namespace = repo.GenerateNamespace(owner, repoName)
	}

	// Check if namespace exists
	exists, err := store.NamespaceExists(namespace)
	if err != nil {
		return fmt.Errorf("check namespace: %w", err)
	}

	if exists {
		fmt.Printf("Namespace '%s' already exists.\n", namespace)
		fmt.Print("Enter alternative namespace: ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}

		namespace = strings.TrimSpace(input)
		if namespace == "" {
			return errors.New("namespace cannot be empty")
		}

		// Check again
		exists, err = store.NamespaceExists(namespace)
		if err != nil {
			return fmt.Errorf("check namespace: %w", err)
		}
		if exists {
			return fmt.Errorf("namespace '%s' already exists", namespace)
		}
	}

	fmt.Printf("Registering %s...\n", url)

	config, err := store.Add(url, namespace)
	if err != nil {
		if errors.Is(err, repo.ErrNamespaceExists) {
			return fmt.Errorf("namespace '%s' already exists", namespace)
		}
		return fmt.Errorf("add repository: %w", err)
	}

	fmt.Printf("Repository registered successfully!\n")
	fmt.Printf("  Namespace:      %s\n", config.Namespace)
	fmt.Printf("  URL:            %s\n", config.URL)
	fmt.Printf("  Default Branch: %s\n", config.DefaultBranch)
	fmt.Println()
	fmt.Printf("Browse packages: jd pkg browse %s\n", config.Namespace)

	return nil
}
