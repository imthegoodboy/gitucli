package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/parth/gitucli/internal/core"
	"github.com/parth/gitucli/internal/storage"
	"github.com/parth/gitucli/internal/textui"
)

type menuItem struct {
	key         string
	title       string
	description string
	action      func(context.Context, *commandEnv, *bufio.Reader) error
}

func runInteractiveShell(ctx context.Context, env *commandEnv) error {
	reader := bufio.NewReader(env.in)
	for {
		clearScreen(env.out)
		renderInteractiveHome(env.out)
		choice, err := readLine(reader, env.out, "Choose")
		if err != nil {
			return nil
		}
		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "1":
			if err := interactiveInit(ctx, env, reader); err != nil {
				showInteractiveError(env, err)
			}
			pause(reader, env.out)
		case "2":
			if err := interactiveProfiles(ctx, env, reader); err != nil {
				showInteractiveError(env, err)
			}
		case "3":
			if err := interactiveRepoTools(ctx, env, reader); err != nil {
				showInteractiveError(env, err)
			}
		case "4":
			if err := interactiveAutoCommit(ctx, env, reader); err != nil {
				showInteractiveError(env, err)
			}
			pause(reader, env.out)
		case "5":
			if err := interactiveDaemonOnce(ctx, env); err != nil {
				showInteractiveError(env, err)
			}
			pause(reader, env.out)
		case "6":
			renderInteractiveHelp(env.out)
			pause(reader, env.out)
		case "q", "quit", "exit", "0":
			fmt.Fprintln(env.out, textui.Muted("bye"))
			return nil
		default:
			fmt.Fprintln(env.out, textui.Warn("Choose a number from the menu, or q to exit."))
			pause(reader, env.out)
		}
	}
}

func renderInteractiveHome(w io.Writer) {
	fmt.Fprint(w, textui.Dashboard([]textui.MenuRow{
		{Index: "1", Command: "Setup a repo", Description: "Guided profile, SSH key, remote, and hook setup"},
		{Index: "2", Command: "Profiles", Description: "Add, list, show, remove, or generate keys"},
		{Index: "3", Command: "Repo tools", Description: "Validate and repair managed repos"},
		{Index: "4", Command: "Autocommit", Description: "Dry-run, commit now, or schedule checkpoints"},
		{Index: "5", Command: "Daemon sweep", Description: "Check configured repos once"},
		{Index: "6", Command: "Help", Description: "Show docs and command examples"},
		{Index: "0", Command: "Exit", Description: "Close gitu"},
	}))
}

func interactiveInit(ctx context.Context, env *commandEnv, reader *bufio.Reader) error {
	clearScreen(env.out)
	fmt.Fprintln(env.out, textui.Section("Setup A Repo"))
	repoPath := promptDefault(reader, env.out, "Repo directory", ".")
	profileName := promptDefault(reader, env.out, "Profile name", "")
	if strings.TrimSpace(profileName) == "" {
		return fmt.Errorf("profile name is required")
	}

	svc, closeFn, err := openService(ctx, env)
	if err != nil {
		return err
	}
	defer closeFn()

	profile, err := svc.Store.GetProfileByName(ctx, profileName)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return err
		}
		fmt.Fprintln(env.out, textui.Muted("Profile not found. Let's create it first."))
		profile, err = createProfileFromPrompts(ctx, svc, reader, env.out, profileName)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(profile.SSHKeyPath); err != nil {
		if promptYesNo(reader, env.out, "SSH key is missing. Generate it now?", true) {
			if _, err := svc.GenerateKey(ctx, profile.Name, false); err != nil {
				return err
			}
			fmt.Fprintf(env.out, "%s Add this public key to GitHub: %s.pub\n", textui.Success("[OK]"), profile.SSHKeyPath)
		}
	}

	repoSlug := promptDefault(reader, env.out, "GitHub repo slug (owner/name)", "")
	remoteName := promptDefault(reader, env.out, "Remote name", "origin")
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
}

func interactiveProfiles(ctx context.Context, env *commandEnv, reader *bufio.Reader) error {
	for {
		clearScreen(env.out)
		fmt.Fprintln(env.out, textui.Page("Profiles", []textui.MenuRow{
			{Index: "1", Command: "Add / update profile", Description: "Create or change one GitHub identity"},
			{Index: "2", Command: "List profiles", Description: "Show all saved identities"},
			{Index: "3", Command: "Show profile", Description: "Inspect one identity"},
			{Index: "4", Command: "Generate SSH key", Description: "Create a key for a profile"},
			{Index: "5", Command: "Remove profile", Description: "Delete an unused profile"},
			{Index: "0", Command: "Back", Description: "Return to main menu"},
		}))
		choice, err := readLine(reader, env.out, "Choose")
		if err != nil {
			return nil
		}
		switch strings.TrimSpace(choice) {
		case "1":
			if err := interactiveProfileAdd(ctx, env, reader); err != nil {
				showInteractiveError(env, err)
			}
			pause(reader, env.out)
		case "2":
			if err := interactiveProfileList(ctx, env); err != nil {
				showInteractiveError(env, err)
			}
			pause(reader, env.out)
		case "3":
			if err := interactiveProfileShow(ctx, env, reader); err != nil {
				showInteractiveError(env, err)
			}
			pause(reader, env.out)
		case "4":
			if err := interactiveKeyGenerate(ctx, env, reader); err != nil {
				showInteractiveError(env, err)
			}
			pause(reader, env.out)
		case "5":
			if err := interactiveProfileRemove(ctx, env, reader); err != nil {
				showInteractiveError(env, err)
			}
			pause(reader, env.out)
		case "0", "b", "back":
			return nil
		}
	}
}

func interactiveRepoTools(ctx context.Context, env *commandEnv, reader *bufio.Reader) error {
	for {
		clearScreen(env.out)
		fmt.Fprintln(env.out, textui.Page("Repo Tools", []textui.MenuRow{
			{Index: "1", Command: "Validate repo", Description: "Check identity safety"},
			{Index: "2", Command: "Repair repo", Description: "Restore managed settings"},
			{Index: "0", Command: "Back", Description: "Return to main menu"},
		}))
		choice, err := readLine(reader, env.out, "Choose")
		if err != nil {
			return nil
		}
		switch strings.TrimSpace(choice) {
		case "1":
			if err := interactiveValidate(ctx, env, reader); err != nil {
				showInteractiveError(env, err)
			}
			pause(reader, env.out)
		case "2":
			if err := interactiveRepair(ctx, env, reader); err != nil {
				showInteractiveError(env, err)
			}
			pause(reader, env.out)
		case "0", "b", "back":
			return nil
		}
	}
}

func interactiveAutoCommit(ctx context.Context, env *commandEnv, reader *bufio.Reader) error {
	clearScreen(env.out)
	fmt.Fprintln(env.out, textui.Section("Autocommit"))
	repoPath := promptDefault(reader, env.out, "Repo directory", ".")
	message := promptDefault(reader, env.out, "Commit message", "")
	mode := promptDefault(reader, env.out, "Mode: dry-run, now, at, interval", "dry-run")
	push := false
	atClock := ""
	interval := ""

	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "dry-run", "dry", "d":
	case "now", "commit", "n":
		push = promptYesNo(reader, env.out, "Push after commit?", false)
	case "at":
		atClock = promptDefault(reader, env.out, "Start time HH:MM", "22:30")
		push = promptYesNo(reader, env.out, "Push after commit?", false)
	case "interval":
		interval = promptDefault(reader, env.out, "Interval duration", "30m")
		push = promptYesNo(reader, env.out, "Push after each commit?", false)
	default:
		return fmt.Errorf("unknown autocommit mode %q", mode)
	}

	svc, closeFn, err := openService(ctx, env)
	if err != nil {
		return err
	}
	defer closeFn()

	if atClock != "" {
		delay, err := core.NextClockDelay(time.Now(), atClock)
		if err != nil {
			return err
		}
		if err := waitWithSpinner(ctx, env.out, delay, fmt.Sprintf("waiting %s until %s", delay.Round(time.Second), atClock)); err != nil {
			return err
		}
	}

	run := func(dryRun bool) error {
		result, err := svc.AutoCommitOnce(ctx, core.AutoCommitOptions{
			RepoPath: repoPath,
			Message:  message,
			Push:     push && !dryRun,
			DryRun:   dryRun,
		})
		if dryRun {
			printAutoCommitDryRun(env.out, result, err)
		} else {
			printAutoCommitResult(env.out, result, err)
		}
		return err
	}

	if strings.ToLower(strings.TrimSpace(mode)) == "interval" {
		d, err := time.ParseDuration(interval)
		if err != nil {
			return err
		}
		if err := run(false); err != nil {
			return err
		}
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := run(false); err != nil {
					fmt.Fprintln(env.errOut, err)
				}
			}
		}
	}
	return run(strings.HasPrefix(strings.ToLower(strings.TrimSpace(mode)), "dry"))
}

func interactiveDaemonOnce(ctx context.Context, env *commandEnv) error {
	svc, closeFn, err := openService(ctx, env)
	if err != nil {
		return err
	}
	defer closeFn()
	reports, err := svc.DaemonSweep(ctx)
	if err != nil {
		return err
	}
	for _, report := range reports {
		if report.OK {
			fmt.Fprintf(env.out, "%s %s\n", textui.Success("[OK]"), report.RepoPath)
		} else {
			fmt.Fprintf(env.out, "%s %s (%d issue(s))\n", textui.Error("[ISSUES]"), report.RepoPath, len(report.Issues))
		}
	}
	return nil
}

func interactiveProfileAdd(ctx context.Context, env *commandEnv, reader *bufio.Reader) error {
	svc, closeFn, err := openService(ctx, env)
	if err != nil {
		return err
	}
	defer closeFn()
	name := promptDefault(reader, env.out, "Profile name", "")
	p, err := createProfileFromPrompts(ctx, svc, reader, env.out, name)
	if err != nil {
		return err
	}
	fmt.Fprintf(env.out, "%s Saved profile %s with alias %s\n", textui.Success("[OK]"), textui.Accent(p.Name), textui.Command(p.SSHAlias))
	return nil
}

func createProfileFromPrompts(ctx context.Context, svc *core.Service, reader *bufio.Reader, w io.Writer, name string) (storage.Profile, error) {
	name = promptMissing(reader, w, "Profile name", name, "")
	githubUser := promptDefault(reader, w, "GitHub username", name)
	gitName := promptDefault(reader, w, "Git commit name", githubUser)
	email := promptDefault(reader, w, "Git commit email", "")
	keyPath := promptDefault(reader, w, "SSH key path", core.DefaultKeyPath(name))
	alias := promptDefault(reader, w, "SSH alias", core.DefaultAlias(githubUser))
	return svc.SaveProfile(ctx, core.ProfileInput{
		Name:       name,
		GitHubUser: githubUser,
		GitName:    gitName,
		Email:      email,
		SSHKeyPath: keyPath,
		SSHAlias:   alias,
	})
}

func interactiveProfileList(ctx context.Context, env *commandEnv) error {
	svc, closeFn, err := openService(ctx, env)
	if err != nil {
		return err
	}
	defer closeFn()
	profiles, err := svc.Store.ListProfiles(ctx)
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		fmt.Fprintln(env.out, textui.Muted("No profiles yet. Choose Profiles > Add profile."))
		return nil
	}
	for _, p := range profiles {
		fmt.Fprintf(env.out, "%s %s  %s  %s\n", textui.Accent(p.Name), textui.Muted(p.GitHubUser), p.Email, textui.Command(p.SSHAlias))
	}
	return nil
}

func interactiveProfileShow(ctx context.Context, env *commandEnv, reader *bufio.Reader) error {
	name := promptDefault(reader, env.out, "Profile name", "")
	svc, closeFn, err := openService(ctx, env)
	if err != nil {
		return err
	}
	defer closeFn()
	p, err := svc.Store.GetProfileByName(ctx, name)
	if err != nil {
		return err
	}
	fmt.Fprintf(env.out, "%s\n%s\n%s\n%s\n%s\n%s\n",
		textui.KeyValue("Name", textui.Accent(p.Name)),
		textui.KeyValue("GitHub", p.GitHubUser),
		textui.KeyValue("Git name", p.GitName),
		textui.KeyValue("Email", p.Email),
		textui.KeyValue("Alias", textui.Command(p.SSHAlias)),
		textui.KeyValue("Key", p.SSHKeyPath))
	return nil
}

func interactiveKeyGenerate(ctx context.Context, env *commandEnv, reader *bufio.Reader) error {
	name := promptDefault(reader, env.out, "Profile name", "")
	force := promptYesNo(reader, env.out, "Overwrite existing key?", false)
	svc, closeFn, err := openService(ctx, env)
	if err != nil {
		return err
	}
	defer closeFn()
	p, err := svc.GenerateKey(ctx, name, force)
	if err != nil {
		return err
	}
	fmt.Fprintf(env.out, "%s Generated SSH key for %s:\n%s\n%s.pub\n", textui.Success("[OK]"), textui.Accent(p.Name), p.SSHKeyPath, p.SSHKeyPath)
	return nil
}

func interactiveProfileRemove(ctx context.Context, env *commandEnv, reader *bufio.Reader) error {
	name := promptDefault(reader, env.out, "Profile name", "")
	if !promptYesNo(reader, env.out, "Remove this profile?", false) {
		return nil
	}
	svc, closeFn, err := openService(ctx, env)
	if err != nil {
		return err
	}
	defer closeFn()
	if err := svc.Store.DeleteProfile(ctx, name); err != nil {
		return err
	}
	if err := svc.SyncSSHConfig(ctx); err != nil {
		return err
	}
	fmt.Fprintf(env.out, "%s Removed profile %s\n", textui.Success("[OK]"), textui.Accent(name))
	return nil
}

func interactiveValidate(ctx context.Context, env *commandEnv, reader *bufio.Reader) error {
	path := promptDefault(reader, env.out, "Repo directory", ".")
	svc, closeFn, err := openService(ctx, env)
	if err != nil {
		return err
	}
	defer closeFn()
	report, err := svc.ValidateRepo(ctx, path)
	if err != nil {
		return err
	}
	printReport(env.out, report)
	if !report.OK {
		return fmt.Errorf("repo is not identity-safe; run gitu repair %q", report.RepoPath)
	}
	return nil
}

func interactiveRepair(ctx context.Context, env *commandEnv, reader *bufio.Reader) error {
	path := promptDefault(reader, env.out, "Repo directory", ".")
	svc, closeFn, err := openService(ctx, env)
	if err != nil {
		return err
	}
	defer closeFn()
	report, err := svc.RepairRepo(ctx, path)
	if err != nil {
		return err
	}
	printReport(env.out, report)
	return nil
}

func renderInteractiveHelp(w io.Writer) {
	clearScreen(w)
	fmt.Fprintln(w, textui.Section("Help"))
	fmt.Fprintln(w, textui.KeyValue("User guide", "docs/USER_GUIDE.md"))
	fmt.Fprintln(w, textui.KeyValue("Autocommit", "docs/AUTOCOMMIT.md"))
	fmt.Fprintln(w, textui.KeyValue("Identity model", "docs/IDENTITY_MODEL.md"))
	fmt.Fprintln(w)
	fmt.Fprintln(w, textui.Command("gitu init --help"))
	fmt.Fprintln(w, textui.Command("gitu autocommit --help"))
	fmt.Fprintln(w, textui.Command("gitu validate --help"))
}

func readLine(reader *bufio.Reader, w io.Writer, label string) (string, error) {
	fmt.Fprintf(w, "%s %s ", textui.Accent(label), textui.Muted(">"))
	text, err := reader.ReadString('\n')
	if err != nil && len(text) == 0 {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func pause(reader *bufio.Reader, w io.Writer) {
	fmt.Fprintf(w, "\n%s ", textui.Muted("Press Enter to continue..."))
	_, _ = reader.ReadString('\n')
}

func clearScreen(w io.Writer) {
	fmt.Fprint(w, "\033[H\033[2J")
}

func showInteractiveError(env *commandEnv, err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(env.errOut, formatCLIError(err))
}
