package x_log

import (
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap/zapcore"
)

// Styles defines visual styles for log components.
type Styles struct {
	Levels map[zapcore.Level]lipgloss.Style // log levels: DEB, INF, WAR, ERR
	Keys   map[string]lipgloss.Style        // keys in structured fields
	Values map[string]lipgloss.Style        // values in structured fields
	Time   lipgloss.Style                   // timestamp style
}

// DefaultStyles returns the default log styling.
func DefaultStyles() *Styles {
	return &Styles{
		Levels: map[zapcore.Level]lipgloss.Style{
			zapcore.DebugLevel: lipgloss.NewStyle().
				SetString("DEB").
				Foreground(lipgloss.Color("#8d8d8d")).
				PaddingLeft(1).PaddingRight(1),

			zapcore.InfoLevel: lipgloss.NewStyle().
				SetString("INF").
				Foreground(lipgloss.Color("#4589ff")).
				PaddingLeft(1).PaddingRight(1),

			zapcore.WarnLevel: lipgloss.NewStyle().
				SetString("WAR").
				Foreground(lipgloss.Color("#ff832b")).
				PaddingLeft(1).PaddingRight(1),

			zapcore.ErrorLevel: lipgloss.NewStyle().
				SetString("ERR").
				Foreground(lipgloss.Color("#fa4d56")).
				PaddingLeft(1).PaddingRight(1),
		},

		Keys: map[string]lipgloss.Style{
			"err":    lipgloss.NewStyle().Foreground(lipgloss.Color("#fa4d56")),
			"user":   lipgloss.NewStyle().Foreground(lipgloss.Color("#a6c8ff")),
			"file":   lipgloss.NewStyle().Foreground(lipgloss.Color("#a6c8ff")),
			"ip":     lipgloss.NewStyle().Foreground(lipgloss.Color("#a6c8ff")),
			"step":   lipgloss.NewStyle().Foreground(lipgloss.Color("#a6c8ff")),
			"module": lipgloss.NewStyle().Foreground(lipgloss.Color("#a6c8ff")),
		},

		Values: map[string]lipgloss.Style{
			"err":    lipgloss.NewStyle().Bold(true),
			"user":   lipgloss.NewStyle().Italic(true),
			"file":   lipgloss.NewStyle().Italic(true),
			"ip":     lipgloss.NewStyle(),
			"step":   lipgloss.NewStyle().Bold(true),
			"module": lipgloss.NewStyle(),
		},

		Time: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#525252")).
			Width(14),
	}
}
