// file:mini/pkg/x_log/style.go
package x_log

import (
	"io"

	"github.com/charmbracelet/lipgloss"
)

//---------------------
// Constants
//---------------------

const (
	ColorTeal40    = "#3ddbd9"
	ColorBlue40    = "#78a9ff"
	ColorBlue60    = "#4589ff"
	ColorBlue70    = "#0043ce"
	ColorBlueBase  = "#0f62fe"
	ColorRed60     = "#da1e28"
	ColorRedStrong = "#ff0000"
	ColorOrange40  = "#ff832b"
	ColorGray60    = "#8d8d8d"
	ColorGray10    = "#f4f4f4"
	ColorGray90    = "#262626"
	ColorGreen40   = "#42be65"
)

//---------------------
// Types
//---------------------

type Styles struct {
	Out               io.Writer                 // reserved output writer
	Timestamp         lipgloss.Style            // timestamp style
	Levels            map[string]lipgloss.Style // level-specific styles
	Keys              map[string]lipgloss.Style // field key styles
	Values            map[string]lipgloss.Style // field value styles
	DefaultKeyStyle   lipgloss.Style            // fallback for unknown keys
	DefaultValueStyle lipgloss.Style            // fallback for unknown values
}

//---------------------
// Default Styles (Dark)
//---------------------

func DefaultStylesDark() *Styles {
	return &Styles{
		Timestamp: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGray60)).
			Width(16),

		DefaultKeyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorBlue40)),

		DefaultValueStyle: lipgloss.NewStyle(),

		Levels: map[string]lipgloss.Style{
			"info":  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen40)),
			"debug": lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray60)),

			"warn":  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOrange40)),
			"error": lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed60)),
			"panic": lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRedStrong)),
		},

		Keys: map[string]lipgloss.Style{
			"user":   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue40)),
			"file":   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue40)),
			"ip":     lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue40)),
			"step":   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue40)),
			"module": lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue40)),
			"err":    lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed60)),
		},

		Values: map[string]lipgloss.Style{
			"user":   lipgloss.NewStyle().Italic(true),
			"file":   lipgloss.NewStyle().Italic(true),
			"step":   lipgloss.NewStyle().Bold(true),
			"err":    lipgloss.NewStyle().Bold(true),
			"module": lipgloss.NewStyle(),
		},
	}
}
