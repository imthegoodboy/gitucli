package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/parth/gitucli/internal/core"
	"github.com/parth/gitucli/internal/storage"
	"github.com/parth/gitucli/internal/textui"
	"github.com/spf13/cobra"
)

type commandEnv struct {
	dbPath string
	in     io.Reader
	out    io.Writer
	errOut io.Writer
}

func Execute() error {
	env := &commandEnv{
		in:     os.Stdin,
		out:    os.Stdout,
		errOut: os.Stderr,
	}
	return newRootCommand(env).Execute()
}

func newRootCommand(env *commandEnv) *cobra.Command {
	root := &cobra.Command{
		Use:   "gitu",
		Short: "Multi GitHub identity manager",
		Long:  "gituCli isolates GitHub identities per repository using local Git config, SSH aliases, remotes, and hooks.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return textui.Render(env.out)
		},
	}
	root.PersistentFlags().StringVar(&env.dbPath, "db", "", "override gitu SQLite database path")

	root.AddCommand(newInitCommand(env))
	root.AddCommand(newProfileCommand(env))
	root.AddCommand(newKeyCommand(env))
	root.AddCommand(newValidateCommand(env))
	root.AddCommand(newRepairCommand(env))
	root.AddCommand(newGuardCommand(env))
	root.AddCommand(newDaemonCommand(env))
	return root
}

func newInitCommand(env *commandEnv) *cobra.Command {
	var profileName, githubUser, gitName, email, keyPath, alias, remoteName, repoSlug string
	var generateKey bool

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a repo with one GitHub identity",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			repoPath := "."
			if len(args) == 1 {
				repoPath = args[0]
			}
			if remoteName == "" {
				remoteName = "origin"
			}

			svc, closeFn, err := openService(ctx, env)
			if err != nil {
				return err
			}
			defer closeFn()

			reader := bufio.NewReader(env.in)
			profileName = promptMissing(reader, env.out, "Profile name", profileName, "")
			if strings.TrimSpace(profileName) == "" {
				return fmt.Errorf("profile name is required")
			}

			profile, err := svc.Store.GetProfileByName(ctx, profileName)
			if err != nil {
				if !errors.Is(err, storage.ErrNotFound) {
					return err
				}
				githubUser = promptMissing(reader, env.out, "GitHub username", githubUser, profileName)
				gitName = promptMissing(reader, env.out, "Git commit name", gitName, githubUser)
				email = promptMissing(reader, env.out, "Git commit email", email, "")
				keyPath = promptMissing(reader, env.out, "SSH key path", keyPath, core.DefaultKeyPath(profileName))
				alias = promptMissing(reader, env.out, "SSH alias", alias, core.DefaultAlias(githubUser))
				profile, err = svc.SaveProfile(ctx, core.ProfileInput{
					Name:       profileName,
					GitHubUser: githubUser,
					GitName:    gitName,
					Email:      email,
					SSHKeyPath: keyPath,
					SSHAlias:   alias,
				})
				if err != nil {
					return err
				}
			}

			if _, err := os.Stat(profile.SSHKeyPath); err != nil {
				shouldGenerate := generateKey || promptYesNo(reader, env.out, fmt.Sprintf("SSH key missing at %s. Generate it now?", profile.SSHKeyPath), true)
				if shouldGenerate {
					if _, err := svc.GenerateKey(ctx, profile.Name, false); err != nil {
						return err
					}
					fmt.Fprintf(env.out, "Generated SSH key. Add this public key to GitHub account %s:\n%s.pub\n\n", profile.GitHubUser, profile.SSHKeyPath)
				}
			}

			if repoSlug == "" {
				repoSlug = promptMissing(reader, env.out, "GitHub repo slug if no origin remote exists (owner/name)", repoSlug, "")
			}
			report, err := svc.ConfigureRepo(ctx, core.InitOptions{
				RepoPath:    repoPath,
				ProfileName: profile.Name,
				RemoteName:  remoteName,
				RepoSlug:    repoSlug,
			})
			if err != nil {
				return err
			}
			printReport(env.out, report)
			return nil
		},
	}

	cmd.Flags().StringVar(&profileName, "profile", "", "profile name to use")
	cmd.Flags().StringVar(&githubUser, "github-user", "", "GitHub username for a new profile")
	cmd.Flags().StringVar(&gitName, "git-name", "", "Git author name for a new profile")
	cmd.Flags().StringVar(&email, "email", "", "Git author email for a new profile")
	cmd.Flags().StringVar(&keyPath, "key", "", "SSH private key path for a new profile")
	cmd.Flags().StringVar(&alias, "alias", "", "SSH host alias for a new profile")
	cmd.Flags().StringVar(&remoteName, "remote", "origin", "remote name to manage")
	cmd.Flags().StringVar(&repoSlug, "repo", "", "GitHub repo slug owner/name when no remote exists")
	cmd.Flags().BoolVar(&generateKey, "generate-key", false, "generate missing SSH key without prompting")
	return cmd
}

func newProfileCommand(env *commandEnv) *cobra.Command {
	cmd := &cobra.Command{Use: "profile", Short: "Manage identity profiles"}
	cmd.AddCommand(newProfileAddCommand(env))
	cmd.AddCommand(newProfileListCommand(env))
	cmd.AddCommand(newProfileShowCommand(env))
	cmd.AddCommand(newProfileRemoveCommand(env))
	return cmd
}

func newProfileAddCommand(env *commandEnv) *cobra.Command {
	var name, githubUser, gitName, email, keyPath, alias string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add or update an identity profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			reader := bufio.NewReader(env.in)
			name = promptMissing(reader, env.out, "Profile name", name, "")
			githubUser = promptMissing(reader, env.out, "GitHub username", githubUser, name)
			gitName = promptMissing(reader, env.out, "Git commit name", gitName, githubUser)
			email = promptMissing(reader, env.out, "Git commit email", email, "")
			keyPath = promptMissing(reader, env.out, "SSH key path", keyPath, core.DefaultKeyPath(name))
			alias = promptMissing(reader, env.out, "SSH alias", alias, core.DefaultAlias(githubUser))

			svc, closeFn, err := openService(ctx, env)
			if err != nil {
				return err
			}
			defer closeFn()

			p, err := svc.SaveProfile(ctx, core.ProfileInput{
				Name:       name,
				GitHubUser: githubUser,
				GitName:    gitName,
				Email:      email,
				SSHKeyPath: keyPath,
				SSHAlias:   alias,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(env.out, "Saved profile %s with alias %s\n", p.Name, p.SSHAlias)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "profile name")
	cmd.Flags().StringVar(&githubUser, "github-user", "", "GitHub username")
	cmd.Flags().StringVar(&gitName, "git-name", "", "Git author name")
	cmd.Flags().StringVar(&email, "email", "", "Git author email")
	cmd.Flags().StringVar(&keyPath, "key", "", "SSH private key path")
	cmd.Flags().StringVar(&alias, "alias", "", "SSH host alias")
	return cmd
}

func newProfileListCommand(env *commandEnv) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, closeFn, err := openService(cmd.Context(), env)
			if err != nil {
				return err
			}
			defer closeFn()
			profiles, err := svc.Store.ListProfiles(cmd.Context())
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(env.out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tGITHUB\tEMAIL\tALIAS\tKEY")
			for _, p := range profiles {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", p.Name, p.GitHubUser, p.Email, p.SSHAlias, p.SSHKeyPath)
			}
			return w.Flush()
		},
	}
}

func newProfileShowCommand(env *commandEnv) *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show one profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, closeFn, err := openService(cmd.Context(), env)
			if err != nil {
				return err
			}
			defer closeFn()
			p, err := svc.Store.GetProfileByName(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(env.out, "Name: %s\nGitHub: %s\nGit name: %s\nEmail: %s\nAlias: %s\nKey: %s\n",
				p.Name, p.GitHubUser, p.GitName, p.Email, p.SSHAlias, p.SSHKeyPath)
			return nil
		},
	}
}

func newProfileRemoveCommand(env *commandEnv) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an unused profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, closeFn, err := openService(cmd.Context(), env)
			if err != nil {
				return err
			}
			defer closeFn()
			if err := svc.Store.DeleteProfile(cmd.Context(), args[0]); err != nil {
				return err
			}
			if err := svc.SyncSSHConfig(cmd.Context()); err != nil {
				return err
			}
			fmt.Fprintf(env.out, "Removed profile %s\n", args[0])
			return nil
		},
	}
}

func newKeyCommand(env *commandEnv) *cobra.Command {
	cmd := &cobra.Command{Use: "key", Short: "Manage SSH keys"}
	var force bool
	gen := &cobra.Command{
		Use:   "generate <profile>",
		Short: "Generate an ed25519 SSH key for a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, closeFn, err := openService(cmd.Context(), env)
			if err != nil {
				return err
			}
			defer closeFn()
			p, err := svc.GenerateKey(cmd.Context(), args[0], force)
			if err != nil {
				return err
			}
			fmt.Fprintf(env.out, "Generated SSH key for %s:\n%s\n%s.pub\n", p.Name, p.SSHKeyPath, p.SSHKeyPath)
			return nil
		},
	}
	gen.Flags().BoolVar(&force, "force", false, "overwrite an existing key")
	cmd.AddCommand(gen)
	return cmd
}

func newValidateCommand(env *commandEnv) *cobra.Command {
	return &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate repo identity safety",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) == 1 {
				path = args[0]
			}
			svc, closeFn, err := openService(cmd.Context(), env)
			if err != nil {
				return err
			}
			defer closeFn()
			report, err := svc.ValidateRepo(cmd.Context(), path)
			if err != nil {
				return err
			}
			printReport(env.out, report)
			if !report.OK {
				return fmt.Errorf("repo is not identity-safe; run gitu repair %q", report.RepoPath)
			}
			return nil
		},
	}
}

func newRepairCommand(env *commandEnv) *cobra.Command {
	return &cobra.Command{
		Use:   "repair [path]",
		Short: "Repair managed repo identity settings",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) == 1 {
				path = args[0]
			}
			svc, closeFn, err := openService(cmd.Context(), env)
			if err != nil {
				return err
			}
			defer closeFn()
			report, err := svc.RepairRepo(cmd.Context(), path)
			if err != nil {
				return err
			}
			printReport(env.out, report)
			if !report.OK {
				return fmt.Errorf("repo still has identity issues")
			}
			return nil
		},
	}
}

func newGuardCommand(env *commandEnv) *cobra.Command {
	var repoPath string
	cmd := &cobra.Command{
		Use:   "guard <pre-commit|pre-push>",
		Short: "Internal strict identity guard used by Git hooks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if repoPath == "" {
				repoPath = "."
			}
			svc, closeFn, err := openService(cmd.Context(), env)
			if err != nil {
				return err
			}
			defer closeFn()
			report, err := svc.ValidateRepo(cmd.Context(), repoPath)
			if err != nil {
				return err
			}
			if report.OK {
				return nil
			}
			fmt.Fprintln(env.errOut, "gitu blocked this Git operation because the repo identity is unsafe:")
			for _, issue := range report.Issues {
				fmt.Fprintf(env.errOut, "- %s\n", issue.Message)
			}
			fmt.Fprintf(env.errOut, "Run: gitu repair %q\n", report.RepoPath)
			return fmt.Errorf("identity guard failed for %s", args[0])
		},
	}
	cmd.Flags().StringVar(&repoPath, "repo", "", "repo path")
	return cmd
}

func newDaemonCommand(env *commandEnv) *cobra.Command {
	var interval time.Duration
	var once bool
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Watch configured repos and restore managed hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, closeFn, err := openService(cmd.Context(), env)
			if err != nil {
				return err
			}
			defer closeFn()

			run := func() error {
				reports, err := svc.DaemonSweep(cmd.Context())
				if err != nil {
					return err
				}
				for _, report := range reports {
					if report.OK {
						fmt.Fprintf(env.out, "OK %s\n", report.RepoPath)
					} else {
						fmt.Fprintf(env.out, "ISSUES %s (%d)\n", report.RepoPath, len(report.Issues))
					}
				}
				return nil
			}
			if once {
				return run()
			}
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			if err := run(); err != nil {
				return err
			}
			for {
				select {
				case <-cmd.Context().Done():
					return nil
				case <-ticker.C:
					if err := run(); err != nil {
						fmt.Fprintln(env.errOut, err)
					}
				}
			}
		},
	}
	cmd.Flags().DurationVar(&interval, "interval", 30*time.Second, "validation interval")
	cmd.Flags().BoolVar(&once, "once", false, "run one sweep and exit")
	return cmd
}

func openService(ctx context.Context, env *commandEnv) (*core.Service, func(), error) {
	dbPath := env.dbPath
	var err error
	if dbPath == "" {
		dbPath, err = storage.DefaultDBPath()
		if err != nil {
			return nil, nil, err
		}
	}
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, nil, err
	}
	svc, err := core.NewService(store)
	if err != nil {
		_ = store.Close()
		return nil, nil, err
	}
	return svc, func() { _ = store.Close() }, nil
}

func printReport(w io.Writer, report core.Report) {
	if report.OK {
		fmt.Fprintf(w, "OK: %s is identity-safe", report.RepoPath)
		if report.ProfileName != "" {
			fmt.Fprintf(w, " for profile %s", report.ProfileName)
		}
		fmt.Fprintln(w)
		return
	}
	fmt.Fprintf(w, "Issues for %s:\n", report.RepoPath)
	for _, issue := range report.Issues {
		repair := ""
		if issue.Repairable {
			repair = " (repairable)"
		}
		fmt.Fprintf(w, "- [%s] %s%s\n", issue.Severity, issue.Message, repair)
	}
}

func promptDefault(r *bufio.Reader, w io.Writer, label, def string) string {
	if strings.TrimSpace(def) != "" {
		fmt.Fprintf(w, "%s [%s]: ", label, def)
	} else {
		fmt.Fprintf(w, "%s: ", label)
	}
	text, _ := r.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return def
	}
	return text
}

func promptMissing(r *bufio.Reader, w io.Writer, label, value, def string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return promptDefault(r, w, label, def)
}

func promptYesNo(r *bufio.Reader, w io.Writer, label string, def bool) bool {
	suffix := "Y/n"
	if !def {
		suffix = "y/N"
	}
	fmt.Fprintf(w, "%s [%s]: ", label, suffix)
	text, _ := r.ReadString('\n')
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return def
	}
	return text == "y" || text == "yes"
}

func defaultString(value, def string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return def
}
