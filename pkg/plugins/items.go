package plugins

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Items hold the menu items for a plugin.
type Items struct {
	// CycleItems are the items that appear in the menu bar.
	CycleItems []*Item
	// ExpandedItems are the items that appear when the menu
	// is open.
	ExpandedItems []*Item
}

// Item is a single menu item.
type Item struct {
	// Plugin is the Plugin that this item belong to.
	Plugin *Plugin
	// Text is the content of the menu item.
	Text string
	// Params are the parameters associated with this Item.
	Params ItemParams
	// Items are a collection of items that appear as a
	// sub menu.
	Items []*Item
	// Alternate is an Item to be used in place of this when
	// the option key is held by the user.
	// These are added to the previous line when printed with the alternate=true
	// parameter.
	Alternate *Item
}

// DisplayText gets the text that should be displayed for
// this item.
// It takes into account the Length parameter.
func (i Item) DisplayText() string {
	return truncate(i.Text, i.Params.Length)
}

// ItemParams represent parameters for an Item.
type ItemParams struct {
	// Disabled indicates that this Item should appear
	// disabled.
	Disabled bool
	// Separator indicates that this Item is a separator.
	Separator bool
	// Href is the URL to open when the item is clicked.
	Href string
	// Key is the accelerator (shortcut) for this item.
	Key string
	// Color is the color of the text.
	Color string
	// Font is the font for the text.
	Font string
	// Size is the font size.
	Size int
	// Shell is a shell script to run when the item is clicked.
	Shell string
	// ShellParams are the arguments to pass to the shell executable.
	ShellParams []string
	// Terminal indicates whether to run the shell command in a terminal or not.
	// Default is false.
	Terminal bool
	// Refresh indicates whether clicking this item will cause the plugin
	// to refresh or not.
	Refresh bool
	// Dropdown indicates whether the item appears in the dropdown
	// or not.
	Dropdown bool
	// Length is the maximum length of the item before the text will
	// be truncated.
	Length int
	// Trim indicates whether to trim whitespace from the text or not.
	Trim bool
	// Alternate indicates that this item is an alternative for the
	// previous item. It will be shown when the option key is depressed.
	Alternate bool
	// TemplateImage is the te`mplate image for this item.
	TemplateImage string
	// Image is the item for this item.
	Image string
	// Emojize indicates whether to process emoji strings (like :mushroom:)
	// or not.
	Emojize bool
	// ANSI indicates whether to parsing ANSI codes.
	ANSI bool
}

// parseParams parses the parameters from a single line.
// The string without parameters is returned, along with the
// typed ItemParams.
func parseParams(s string) (string, ItemParams, error) {
	params := defaultParams
	pipeIndex := strings.Index(s, "|")
	if pipeIndex < 0 { // no params
		return s, params, nil
	}
	text := s[:pipeIndex]
	paramStr := s[pipeIndex+1:]
	if err := parseParamStr(&params, paramStr); err != nil {
		return text, params, err
	}
	return text, params, nil
}

// parseParamStr parses the parameter string, updating params.
func parseParamStr(params *ItemParams, s string) error {
	var splitStr, endStr string
	for {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			return nil
		}
		splitStr = `=`
		endStr = ` `
		i := strings.Index(s, splitStr)
		if i < 0 {
			return errors.New("malformed parameters: missing equals")
		}
		if len(s) > i+1 && (s[i+1] == '"' || s[i+1] == '\'') {
			// quotes
			endStr = string(s[i+1])
			splitStr = "=" + endStr
		}
		offset := i + len(splitStr)
		key := s[:i]
		if key[0] == '|' {
			key = key[1:]
			key = strings.TrimSpace(key)
		}
		valuePart := s[offset:]
		end := strings.IndexAny(valuePart, endStr+"|")
		if end < 0 {
			end = len(s)
		} else {
			end += offset
		}
		value := s[i+len(splitStr) : end]
		if err := params.setValueByKey(key, value); err != nil {
			return err
		}
		if end+1 > len(s) {
			return nil
		}
		s = s[end+1:]
	}
}

// defaultParams are the default ItemParams.
var defaultParams = ItemParams{
	Dropdown: true,
	Trim:     true,
	Emojize:  true,
	ANSI:     true,
}

func (p *ItemParams) setValueByKey(key, value string) error {
	switch key {
	case "disabled":
		var err error
		p.Disabled, err = parseBool(value)
		if err != nil {
			return errors.Wrap(err, key)
		}
	case "key":
		p.Key = value
	case "href":
		p.Href = value
	case "color":
		p.Color = value
	case "font":
		p.Font = value
	case "size":
		val, err := parseInt(value)
		if err != nil {
			return errors.Wrap(err, key)
		}
		p.Size = val
	case "shell", "bash":
		p.Shell = value
	case "templateImage":
		p.TemplateImage = value
	case "image":
		p.Image = value
	case "terminal":
		var err error
		p.Terminal, err = parseBool(value)
		if err != nil {
			return errors.Wrap(err, key)
		}
	case "refresh":
		var err error
		p.Refresh, err = parseBool(value)
		if err != nil {
			return errors.Wrap(err, key)
		}
	case "dropdown":
		var err error
		p.Dropdown, err = parseBool(value)
		if err != nil {
			return errors.Wrap(err, key)
		}
	case "length":
		val, err := parseInt(value)
		if err != nil {
			return errors.Wrap(err, key)
		}
		p.Length = val
	case "trim":
		var err error
		p.Trim, err = parseBool(value)
		if err != nil {
			return errors.Wrap(err, key)
		}
	case "alternate":
		var err error
		p.Alternate, err = parseBool(value)
		if err != nil {
			return errors.Wrap(err, key)
		}
	case "emojize":
		var err error
		p.Emojize, err = parseBool(value)
		if err != nil {
			return errors.Wrap(err, key)
		}
	case "ansi":
		var err error
		p.ANSI, err = parseBool(value)
		if err != nil {
			return errors.Wrap(err, key)
		}
	default:
		if strings.HasPrefix(key, "param") {
			paramIndex, err := strconv.Atoi(key[5:])
			if err != nil {
				return errors.Errorf("bad parameter: %s (should be paramN)", key)
			}
			for len(p.ShellParams) < paramIndex {
				// ensure the slice is big enough
				p.ShellParams = append(p.ShellParams, "")
			}
			p.ShellParams[paramIndex-1] = value
			return nil
		}
		return errors.Errorf("unknown parameter: %s", key)
	}
	return nil
}

// parseBool parses a boolean from a string (either `true` or `false`),
// returning a nice error if it fails.
func parseBool(s string) (bool, error) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false, errors.Errorf(`expected "true" or "false", not "%s"`, s)
	}
	return b, nil
}

// parseInt parses an int from a string, returning a nice
// error if it fails.
func parseInt(s string) (int, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.Errorf(`expected an int, not "%s"`, s)
	}
	return int(i), nil
}

// truncate shrinks a string if it's too long.
func truncate(s string, max int) string {
	runes := []rune(s)
	if max > 0 && len(runes) > max {
		s = string(runes[:max-1]) + "…"
		return s
	}
	return s
}
