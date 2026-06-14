package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Options controls CLI presentation.
type Options struct {
	Quiet   bool
	Verbose bool
}

// Default is the active presentation config (set from cobra flags).
var Default Options

var (
	out io.Writer = os.Stdout
	errOut        = os.Stderr

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	promptStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	cmdStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(14)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 2)

	doneBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("236")).
			Padding(1, 2)
)

func colorEnabled() bool {
	if Default.Quiet || os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func render(style lipgloss.Style, s string) string {
	if !colorEnabled() {
		return s
	}
	return style.Render(s)
}

// Blank prints a blank line unless quiet.
func Blank() {
	if Default.Quiet {
		return
	}
	fmt.Fprintln(out)
}

// Line prints a line to stdout.
func Line(s string) {
	if Default.Quiet {
		return
	}
	fmt.Fprintln(out, s)
}

// Muted prints secondary copy.
func Muted(s string) {
	Line(render(mutedStyle, s))
}

// Action prints a status line (e.g. opening browser).
func Action(s string) {
	Line("  " + render(mutedStyle, "→") + " " + s)
}

// Success prints a checkmarked success line.
func Success(s string) {
	if Default.Quiet {
		return
	}
	fmt.Fprintln(out, "  "+render(successStyle, "✓")+" "+s)
}

// Step prints a step header box.
func Step(n, total int, title string) {
	if Default.Quiet {
		return
	}
	Blank()
	header := fmt.Sprintf("Step %d of %d · %s", n, total, title)
	fmt.Fprintln(out, boxStyle.Render(render(titleStyle, header)))
	Blank()
}

// Heading prints a titled box (no step number).
func Heading(title string) {
	if Default.Quiet {
		return
	}
	Blank()
	fmt.Fprintln(out, boxStyle.Render(render(titleStyle, title)))
	Blank()
}

// SetupIntro prints the onboarding welcome once.
func SetupIntro() {
	if Default.Quiet {
		return
	}
	Blank()
	fmt.Fprintln(out, boxStyle.Render(render(titleStyle, "theirtime setup")))
	Blank()
}

// Prompt prints a styled input label (no newline).
func Prompt(label string) {
	if Default.Quiet {
		fmt.Fprint(out, label+": ")
		return
	}
	fmt.Fprint(out, "  "+render(labelStyle, label)+" "+render(promptStyle, "›")+" ")
}

// URL prints a manual-open URL hint on stderr.
func URL(label, url string) {
	fmt.Fprintln(errOut, label)
	fmt.Fprintln(errOut, url)
}

// Command prints an indented copy-paste command.
func Command(name string, args ...string) {
	if Default.Quiet {
		return
	}
	cmd := name
	if len(args) > 0 {
		cmd += " " + strings.Join(args, " ")
	}
	fmt.Fprintln(out, "    "+render(cmdStyle, cmd))
}

// InstallAgentsDone prints success after the menu bar LaunchAgent is running.
func InstallAgentsDone(teammateCount int) {
	if Default.Quiet {
		return
	}
	Blank()
	var b strings.Builder
	b.WriteString(render(titleStyle, "Menu bar agent running"))
	b.WriteString("\n\n")
	b.WriteString("  " + render(successStyle, "✓") + " " + render(mutedStyle, "dev.theirtime.menubar started") + "\n")
	b.WriteString("  " + render(mutedStyle, fmt.Sprintf("Watching %d teammate(s)", teammateCount)) + "\n\n")
	b.WriteString("  " + render(mutedStyle, "Look for avatars in your menu bar — may take a few seconds.") + "\n\n")
	b.WriteString("  " + render(mutedStyle, "Logs:") + "\n")
	b.WriteString("    " + render(cmdStyle, "~/Library/Logs/theirtime/menubar.log"))
	fmt.Fprintln(out, doneBoxStyle.Render(b.String()))
	Blank()
}

// InstallAgentsEmpty prints guidance when no teammates are configured.
func InstallAgentsEmpty() {
	if Default.Quiet {
		return
	}
	Blank()
	var b strings.Builder
	b.WriteString(render(titleStyle, "No teammates yet"))
	b.WriteString("\n\n")
	b.WriteString("  " + render(mutedStyle, "The menu bar agent starts once you add someone to watch.") + "\n\n")
	b.WriteString("  Add a teammate:\n")
	b.WriteString("    " + render(cmdStyle, "theirtime team add bob U012ABCDEF") + "\n\n")
	b.WriteString("  Then run:\n")
	b.WriteString("    " + render(cmdStyle, "theirtime install-agents"))
	fmt.Fprintln(out, doneBoxStyle.Render(b.String()))
	Blank()
}

// DoneCard prints the post-onboard success card with next steps.
func DoneCard() {
	if Default.Quiet {
		return
	}
	Blank()
	var b strings.Builder
	b.WriteString(render(titleStyle, "You're all set"))
	b.WriteString("\n\n")
	b.WriteString("  " + render(mutedStyle, "Find a member ID in Slack:") + "\n")
	b.WriteString("    " + render(mutedStyle, "1. Click a teammate's name → View full profile") + "\n")
	b.WriteString("    " + render(mutedStyle, "2. Click ⋮ (More) → Copy member ID") + "\n")
	b.WriteString("    " + render(mutedStyle, "3. Looks like U012ABCDEF") + "\n\n")
	b.WriteString("  Add a teammate:\n")
	b.WriteString("    " + render(cmdStyle, "theirtime team add bob U012ABCDEF") + "\n\n")
	b.WriteString("  Start the menu bar:\n")
	b.WriteString("    " + render(cmdStyle, "theirtime install-agents") + "\n\n")
	b.WriteString("  Preview without Slack:\n")
	b.WriteString("    " + render(cmdStyle, "theirtime menubar --demo"))
	fmt.Fprintln(out, doneBoxStyle.Render(b.String()))
	Blank()
}

// BrowserSteps prints detailed browser instructions when verbose.
func BrowserSteps(credentialHint string) {
	if Default.Quiet {
		return
	}
	Muted("When the app is created, open App Credentials and paste below:")
	Muted("  " + credentialHint)
	if !Default.Verbose {
		return
	}
	Blank()
	Muted("In the browser:")
	Muted("  1. Sign in to Slack (if asked)")
	Muted("  2. Pick a workspace and click Next")
	Muted("  3. Review the manifest and click Create")
	Muted("  4. Copy Client ID and Client Secret")
	Blank()
}
