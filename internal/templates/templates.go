package templates

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	configs "tektmud/internal/config"
	"tektmud/internal/logger"
	"text/template"
)

// ColorTemplate wraps template with color and i18n support
type ColorTemplate struct {
	tmpl *template.Template
}

type TemplateManager struct {
	templates map[string]*ColorTemplate
}

func NewColorTemplate(name string) *ColorTemplate {
	ct := &ColorTemplate{
		tmpl: template.New(name),
	}
	//Adds our templating functions
	//i18n support is built in here.
	ct.tmpl.Funcs(functionsMap)

	return ct
}

// Parses the template text
func (ct *ColorTemplate) Parse(text string) error {
	_, err := ct.tmpl.Parse(text)
	return err
}

// Executes the tmeplate, processing color codes, substitudes data for template fields
func (ct *ColorTemplate) Execute(data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := ct.tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return ct.processColors(buf.String(), false), nil
}

// processColors converts MUD color codes to ANSI codes
func (ct *ColorTemplate) processColors(text string, useTrueColor bool) string {
	// First handle escaped dollars ($$)
	text = strings.ReplaceAll(text, "$$", "\x00DOLLAR\x00")

	// Pattern for various MUD color formats:
	// $r, $R, $b, $B, etc. (single letter colors)
	// $1r, $1R, $1b, $1B, etc. (single letter colors for BG)
	// $X123456 (hex foreground)
	// $1X123456 (hex background)
	// $123 (256-color foreground)
	// $1123 (256-color background)
	// $0,$n (reset)
	colorRegex := regexp.MustCompile(`\$(1[rRgGyYbBmMcCwWkKdDiIuUlLsS0n]|[rRgGyYbBmMcCwWkKdDiIuUlLsS0n]|X[0-9a-fA-F]{6}|1X[0-9a-fA-F]{6}|\d{1,3}|1\d{1,3})`)

	rgb_to_xterm256 := func(r, g, b int64) int64 {
		r6 := (r*5 + 127) / 255
		g6 := (g*5 + 127) / 255
		b6 := (b*5 + 127) / 255
		return 16 + (36 * r6) + (6 * g6) + b6
	}

	colorMap := defaultColorMap()

	result := colorRegex.ReplaceAllStringFunc(text, func(match string) string {
		code := match[1:] // Remove the $

		// Reset
		if code == "0" || code == "n" {
			return "\033[0m"
		}

		// Single letter colors
		if len(code) == 1 {
			if ansi, exists := colorMap[code]; exists {
				return ansi
			}
			//Fall through to allow for checks for things like $3 or $12
		}

		// Single letter BG colors
		if strings.HasPrefix(code, "1") && len(code) == 2 {
			if ansi, exists := colorMap[code[1:]]; exists {
				return fmt.Sprintf("\033[48;5;%s", ansi[4:])

			}
			//Fall through to allow for checks for things like $3 or $12
		}
		// Hex colors
		if strings.HasPrefix(code, "X") && len(code) == 7 {
			hex := code[1:]
			r, _ := strconv.ParseInt(hex[0:2], 16, 64)
			g, _ := strconv.ParseInt(hex[2:4], 16, 64)
			b, _ := strconv.ParseInt(hex[4:6], 16, 64)
			if useTrueColor {
				return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
			} else {
				//downgrade to 256
				return fmt.Sprintf("\033[38;5;%dm", rgb_to_xterm256(r, g, b))
			}
		}

		// Background hex colors
		if strings.HasPrefix(code, "1X") && len(code) == 8 {
			hex := code[2:]
			r, _ := strconv.ParseInt(hex[0:2], 16, 64)
			g, _ := strconv.ParseInt(hex[2:4], 16, 64)
			b, _ := strconv.ParseInt(hex[4:6], 16, 64)
			if useTrueColor {
				return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
			} else {
				return fmt.Sprintf("\033[48;5;%dm", rgb_to_xterm256(r, g, b))
			}
		}

		// 256-color codes
		colorNum, err := strconv.Atoi(code)
		if err == nil && colorNum >= 0 && colorNum <= 255 {
			return fmt.Sprintf("\033[38;5;%dm", colorNum)
		}

		// Background 256-color codes (1NNN format)
		if strings.HasPrefix(code, "1") && len(code) >= 2 {
			bgCode := code[1:]
			colorNum, err := strconv.Atoi(bgCode)
			if err == nil && colorNum >= 0 && colorNum <= 255 {
				return fmt.Sprintf("\033[48;5;%dm", colorNum)
			}
		}

		return match
	})

	// Restore escaped dollars
	return strings.ReplaceAll(result, "\x00DOLLAR\x00", "$")
}

// defaultColorMap returns MUD-style color code mappings
func defaultColorMap() map[string]string {
	return map[string]string{
		// Standard MUD colors (lowercase = normal, uppercase = bright)
		"k": "\033[30m", // black
		"r": "\033[31m", // red
		"g": "\033[32m", // green
		"y": "\033[33m", // yellow
		"b": "\033[34m", // blue
		"m": "\033[35m", // magenta
		"c": "\033[36m", // cyan
		"w": "\033[37m", // white

		// Bright colors
		"K": "\033[90m", // bright black (dark gray)
		"R": "\033[91m", // bright red
		"G": "\033[92m", // bright green
		"Y": "\033[93m", // bright yellow
		"B": "\033[94m", // bright blue
		"M": "\033[95m", // bright magenta
		"C": "\033[96m", // bright cyan
		"W": "\033[97m", // bright white

		// Text styles
		"d": "\033[1m", // bold
		"D": "\033[2m", // dim
		"i": "\033[3m", // italic
		"u": "\033[4m", // underline
		"l": "\033[5m", // blink
		"s": "\033[7m", // reverse
		"S": "\033[9m", // strikethrough
	}
}

func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates: make(map[string]*ColorTemplate),
	}
}

func (tm *TemplateManager) Process(templateName string, maybeData ...any) (string, error) {

	var data any
	if len(maybeData) > 0 {
		data = maybeData[0]
	}

	prefixes := []string{
		"login",
		"rooms", //Only during dev?
	}
	var reload bool = false
	for _, v := range prefixes {
		if !reload {
			reload = strings.HasPrefix(templateName, v)
		}
	}

	if err := tm.LoadTemplate(templateName, reload); err != nil {
		return "[Error loading template]", err
	}

	output, err := tm.Execute(templateName, data)
	if err != nil {
		return "[Error executing template]", err
	}

	return output, nil
}

// This approach assumes we -always- want to cache a template
// that may not scale at some point. May need to handle differently
// or expire cache or something
func (tm *TemplateManager) LoadTemplate(name string, reload ...bool) error {
	var forceReload bool = false
	if len(reload) > 0 {
		forceReload = reload[0]
	}
	var tmpl *ColorTemplate
	_, exists := tm.templates[name]

	if !exists || forceReload {
		c := configs.GetConfig()
		fullPath := filepath.Join(c.Paths.RootDataDir, c.Paths.Templates, name) + `.template`

		fileContents, err := os.ReadFile(fullPath)

		if err != nil {
			logger.Warn("Unable to load template file", "path", fullPath)
			return err
		}
		tmpl = NewColorTemplate(name)
		if err := tmpl.Parse(string(fileContents)); err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		tm.templates[name] = tmpl
		return nil
	}

	return nil
}

// Executes the template and processes color
func (tm *TemplateManager) Execute(name string, data interface{}) (string, error) {
	tmpl, exists := tm.templates[name]
	if !exists {
		return "", fmt.Errorf("template %s not found", name)
	}
	return tmpl.Execute(data)
}
