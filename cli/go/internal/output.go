package core

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// ── Table ────────────────────────────────────────────────

// Table writes a formatted table to stdout using tablewriter.
func Table(header []string, rows [][]string) {
	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader(header)
	t.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.SetBorder(false)
	t.SetTablePadding("\t")
	t.SetNoWhiteSpace(true)
	t.AppendBulk(rows)
	t.Render()
}

// ── JSON ─────────────────────────────────────────────────

// PrintJSON writes v as indented JSON to stdout.
func PrintJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// ── Spinner ──────────────────────────────────────────────

// StartSpinner creates and starts a spinner with the given message.
// Call s.Stop() when done.
func StartSpinner(msg string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = msg + " "
	s.Start()
	return s
}

// ── Color ────────────────────────────────────────────────

// Pre-built color functions for common use cases.
var (
	Info    = color.New(color.FgCyan, color.Bold).SprintFunc()
	Success = color.New(color.FgGreen).SprintFunc()
	Warn    = color.New(color.FgYellow).SprintFunc()
	Error   = color.New(color.FgRed).SprintFunc()
)

// FprintColorized writes colorized output to w (e.g. os.Stdout).
func FprintColorized(w io.Writer, c color.Color, args ...any) {
	c.Fprint(w, args...)
}
