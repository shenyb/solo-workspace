package core

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// ── Table ────────────────────────────────────────────────

const defaultTermWidth = 120

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

func visibleLen(s string) int {
	return runewidth.StringWidth(stripANSI(s))
}

func padRight(s string, width int) string {
	vlen := visibleLen(s)
	if vlen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vlen)
}

func truncateVisible(s string, max int) string {
	if visibleLen(s) <= max {
		return s
	}
	// Collect leading ANSI sequences so colour survives truncation
	var prefix strings.Builder
	rest := s
	for {
		loc := ansiRE.FindStringIndex(rest)
		if loc == nil || loc[0] != 0 {
			break
		}
		prefix.WriteString(rest[loc[0]:loc[1]])
		rest = rest[loc[1]:]
	}
	clean := stripANSI(rest)

	// Truncate by display width, reserving 1 column for ellipsis
	var out strings.Builder
	w := 0
	for _, r := range clean {
		rw := runewidth.RuneWidth(r)
		if w+rw > max-1 {
			break
		}
		out.WriteRune(r)
		w += rw
	}
	return prefix.String() + out.String() + "…"
}

// Table writes a formatted, adaptive-width table to stdout.
// Column widths auto-fit content; when the table is wider than the terminal
// every column is shrunk proportionally.
func Table(header []string, rows [][]string) {
	numCols := len(header)
	if numCols == 0 {
		return
	}

	// ── measure visible widths per column ──
	colWidths := make([]int, numCols)
	for i, h := range header {
		colWidths[i] = visibleLen(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i >= numCols {
				break
			}
			if w := visibleLen(cell); w > colWidths[i] {
				colWidths[i] = w
			}
		}
	}

	// ── cap individual columns ──
	const maxColWidth = 80
	for i := range colWidths {
		if colWidths[i] > maxColWidth {
			colWidths[i] = maxColWidth
		}
	}

	// ── get terminal width ──
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || termWidth <= 0 {
		termWidth = defaultTermWidth
	}

	// ── shrink proportionally if oversize ──
	const sepLen = 3 // " │ "
	totalNeeded := (numCols - 1) * sepLen
	for _, w := range colWidths {
		totalNeeded += w
	}
	if totalNeeded > termWidth {
		available := termWidth - (numCols-1)*sepLen
		total := 0
		for _, w := range colWidths {
			total += w
		}
		if total > 0 {
			for i := range colWidths {
				// ceiling division so short labels like "pending" survive proportional shrink
				colWidths[i] = (colWidths[i]*available + total - 1) / total
				if colWidths[i] < 6 {
					colWidths[i] = 6
				}
			}
		}
	}

	// ── render ──
	var out strings.Builder

	// header
	for i, h := range header {
		if i > 0 {
			out.WriteString(" │ ")
		}
		out.WriteString(padRight(h, colWidths[i]))
	}
	out.WriteByte('\n')

	// header separator
	for i := 0; i < numCols; i++ {
		if i > 0 {
			out.WriteString("─┼─")
		}
		out.WriteString(strings.Repeat("─", colWidths[i]))
	}
	out.WriteByte('\n')

	// data rows
	for _, row := range rows {
		for i := 0; i < numCols; i++ {
			if i > 0 {
				out.WriteString(" │ ")
			}
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			if visibleLen(cell) > colWidths[i] {
				cell = truncateVisible(cell, colWidths[i])
			}
			out.WriteString(padRight(cell, colWidths[i]))
			// Reset ANSI color after every cell to prevent color leakage from
			// truncated colored cells (e.g. TodoStatus "pending"/"done").
			if ansiRE.MatchString(cell) {
				out.WriteString("\x1b[0m")
			}
		}
		out.WriteByte('\n')
		// blank line between rows for readability
		out.WriteByte('\n')
	}

	fmt.Print(out.String())
}

// ── JSON ─────────────────────────────────────────────────

// PrintJSON writes v as indented JSON to stdout.
func PrintJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "json encode error: %v\n", err)
	}
}

// ── Color ────────────────────────────────────────────────

// Pre-built color functions for common use cases.
var (
	Info    = color.New(color.FgCyan, color.Bold).SprintFunc()
	Success = color.New(color.FgGreen).SprintFunc()
	Warn    = color.New(color.FgYellow).SprintFunc()
	Error   = color.New(color.FgRed).SprintFunc()
)

// TodoStatus returns a colorised status label for todo display.
// done → green "done", pending → yellow "pending".
func TodoStatus(done bool) string {
	if done {
		return Success("done")
	}
	return Warn("pending")
}

