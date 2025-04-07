package x_log

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
)

//
// ---------- IBM Carbon Colors ----------

const (
	ColorTeal40    = "#3ddbd9"
	ColorBlue60    = "#4589ff"
	ColorBlue40    = "#78a9ff"
	ColorBlue70    = "#0043ce"
	ColorBlueBase  = "#0f62fe"
	ColorRed60     = "#da1e28"
	ColorRedStrong = "#ff0000"
	ColorOrange40  = "#ff832b"
	ColorGray60    = "#8d8d8d"
	ColorGray10    = "#f4f4f4"
	ColorGray90    = "#262626"
)

//
// ---------- Styles Definition ----------

// Styles defines all formatting styles used for structured output
type Styles struct {
	Out               io.Writer                 // output target
	Timestamp         lipgloss.Style            // style for timestamps
	Levels            map[Level]lipgloss.Style  // level-to-style mapping
	Keys              map[string]lipgloss.Style // custom field keys
	Values            map[string]lipgloss.Style // custom field values
	DefaultKeyStyle   lipgloss.Style            // fallback for unknown keys
	DefaultValueStyle lipgloss.Style            // fallback for unknown values
}

//
// ---------- Theme Selectors ----------

// DefaultStylesByName returns a theme by name ("dark", "light")
func DefaultStylesByName(name string) *Styles {
	switch strings.ToLower(name) {
	case "light":
		return DefaultStylesLight()
	default:
		return DefaultStylesDark()
	}
}

//
// ---------- Console Formatter ----------

// ConsoleWriterWithStyles builds a zerolog.ConsoleWriter with styles
func ConsoleWriterWithStyles(styles *Styles) zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{
		Out:        styles.Out,
		TimeFormat: zerolog.TimeFieldFormat,

		FormatLevel: func(i any) string {
			lvl := strings.ToLower(fmt.Sprint(i))
			var color string

			switch lvl {
			case "debug":
				color = ColorTeal40
			case "info":
				color = ColorBlue60
			case "warn":
				color = ColorOrange40
			case "error":
				color = ColorRed60
			case "fatal":
				color = ColorRedStrong
			default:
				color = ColorGray60
			}

			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color(color)).
				Padding(0, 1).
				Render(strings.ToUpper(lvl[:3]))
		},

		FormatTimestamp: func(i any) string {
			return styles.Timestamp.Render(fmt.Sprintf("[%s]", i))
		},

		FormatFieldName: func(i any) string {
			key := fmt.Sprint(i)
			style, ok := styles.Keys[key]
			if !ok {
				style = styles.DefaultKeyStyle
			}
			eqStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray60))
			return style.Render(key) + eqStyle.Render("=")
		},

		FormatMessage: func(i any) string {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorGray10)).
				Render(fmt.Sprint(i))
		},
	}
}

//
// ---------- Dark Theme ----------

func DefaultStylesDark() *Styles {
	return &Styles{
		Timestamp: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGray60)).
			Width(16),

		DefaultKeyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorBlue40)),

		DefaultValueStyle: lipgloss.NewStyle(),

		Levels: map[Level]lipgloss.Style{
			InfoLevel:  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue60)),
			WarnLevel:  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOrange40)),
			ErrorLevel: lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed60)),
			FatalLevel: lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRedStrong)),
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

//
// ---------- Light Theme ----------

func DefaultStylesLight() *Styles {
	return &Styles{
		Timestamp: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGray60)).
			Width(16),

		DefaultKeyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorBlueBase)),

		DefaultValueStyle: lipgloss.NewStyle(),

		Levels: map[Level]lipgloss.Style{
			InfoLevel:  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue70)),
			WarnLevel:  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOrange40)),
			ErrorLevel: lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed60)),
			FatalLevel: lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRedStrong)),
		},

		Keys: map[string]lipgloss.Style{
			"user":   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlueBase)),
			"file":   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlueBase)),
			"ip":     lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlueBase)),
			"step":   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlueBase)),
			"module": lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlueBase)),
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
