package ui

import (
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type tokenType int

const (
	tokenWord      tokenType = iota
	tokenPriority            // !1, !2, !3
	tokenDueDate             // ^tomorrow, ^friday
	tokenList                // #Personal
	tokenTag                 // %errands
	tokenRecurring           // *weekly
)

type token struct {
	raw  string
	kind tokenType
}

var (
	smartTokenPriorityStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e")) // tokyo night red
	smartTokenDueDateStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a")) // tokyo night green
	smartTokenListStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#bb9af7")) // tokyo night purple
	smartTokenTagStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff")) // tokyo night cyan
	smartTokenRecurringStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68")) // tokyo night yellow
	smartCursorStyle         = lipgloss.NewStyle().Reverse(true)
)

func styleForToken(kind tokenType) lipgloss.Style {
	switch kind {
	case tokenPriority:
		return smartTokenPriorityStyle
	case tokenDueDate:
		return smartTokenDueDateStyle
	case tokenList:
		return smartTokenListStyle
	case tokenTag:
		return smartTokenTagStyle
	case tokenRecurring:
		return smartTokenRecurringStyle
	default:
		return lipgloss.NewStyle()
	}
}

func classifyWord(word []rune) tokenType {
	if len(word) == 0 {
		return tokenWord
	}
	switch word[0] {
	case '!':
		return tokenPriority
	case '^':
		return tokenDueDate
	case '#':
		return tokenList
	case '%':
		return tokenTag
	case '*':
		return tokenRecurring
	default:
		return tokenWord
	}
}

// tokenize splits runes into tokens, preserving spaces as individual tokenWord tokens.
func tokenize(runes []rune) []token {
	var tokens []token
	var word []rune

	flushWord := func() {
		if len(word) > 0 {
			tokens = append(tokens, token{raw: string(word), kind: classifyWord(word)})
			word = nil
		}
	}

	for _, r := range runes {
		if r == ' ' {
			flushWord()
			tokens = append(tokens, token{raw: " ", kind: tokenWord})
		} else {
			word = append(word, r)
		}
	}
	flushWord()

	return tokens
}

// transformForRTM converts the custom syntax to RTM Smart Add format.
// % (tag prefix) is replaced with # since RTM uses # for both tags and lists.
func transformForRTM(input string) string {
	toks := tokenize([]rune(input))
	var sb strings.Builder
	for _, tok := range toks {
		if tok.kind == tokenTag && len(tok.raw) > 0 {
			sb.WriteByte('#')
			sb.WriteString(tok.raw[1:])
		} else {
			sb.WriteString(tok.raw)
		}
	}
	return sb.String()
}

// SmartInput is a text input with per-token syntax highlighting.
type SmartInput struct {
	prompt string
	value  []rune
	cursor int // 0 <= cursor <= len(value)
}

// NewSmartInput creates a SmartInput with the given prompt string.
func NewSmartInput(prompt string) SmartInput {
	return SmartInput{prompt: prompt}
}

// Value returns the current input as a string.
func (s SmartInput) Value() string {
	return string(s.value)
}

// SetValue replaces the input content and moves the cursor to the end.
func (s *SmartInput) SetValue(v string) {
	s.value = []rune(v)
	s.cursor = len(s.value)
}

// CursorEnd moves the cursor to the end of the input.
func (s *SmartInput) CursorEnd() {
	s.cursor = len(s.value)
}

// View renders the input with syntax highlighting and a block cursor.
func (s SmartInput) View() string {
	toks := tokenize(s.value)
	var sb strings.Builder
	sb.WriteString(s.prompt)

	runesConsumed := 0
	cursorWritten := false

	for _, tok := range toks {
		tokRunes := []rune(tok.raw)
		tokLen := len(tokRunes)
		style := styleForToken(tok.kind)

		if !cursorWritten && s.cursor >= runesConsumed && s.cursor < runesConsumed+tokLen {
			localOffset := s.cursor - runesConsumed
			if localOffset > 0 {
				sb.WriteString(style.Render(string(tokRunes[:localOffset])))
			}
			sb.WriteString(smartCursorStyle.Render(string(tokRunes[localOffset : localOffset+1])))
			if localOffset+1 < tokLen {
				sb.WriteString(style.Render(string(tokRunes[localOffset+1:])))
			}
			cursorWritten = true
		} else {
			sb.WriteString(style.Render(tok.raw))
		}

		runesConsumed += tokLen
	}

	if !cursorWritten {
		sb.WriteString(smartCursorStyle.Render(" "))
	}

	return sb.String()
}

// Update handles keypresses. enter and esc are passed through unchanged — the
// caller is responsible for intercepting them.
func (s SmartInput) Update(msg tea.KeyMsg) (SmartInput, tea.Cmd) {
	switch msg.String() {
	case "space":
		s.value = append(s.value[:s.cursor:s.cursor], append([]rune{' '}, s.value[s.cursor:]...)...)
		s.cursor++
	case "backspace":
		if s.cursor > 0 {
			s.value = append(s.value[:s.cursor-1:s.cursor-1], s.value[s.cursor:]...)
			s.cursor--
		}
	case "delete":
		if s.cursor < len(s.value) {
			s.value = append(s.value[:s.cursor:s.cursor], s.value[s.cursor+1:]...)
		}
	case "left", "ctrl+b":
		if s.cursor > 0 {
			s.cursor--
		}
	case "right", "ctrl+f":
		if s.cursor < len(s.value) {
			s.cursor++
		}
	case "home", "ctrl+a":
		s.cursor = 0
	case "end", "ctrl+e":
		s.cursor = len(s.value)
	case "ctrl+w":
		i := s.cursor
		for i > 0 && s.value[i-1] == ' ' {
			i--
		}
		for i > 0 && s.value[i-1] != ' ' {
			i--
		}
		s.value = append(s.value[:i:i], s.value[s.cursor:]...)
		s.cursor = i
	case "ctrl+u":
		s.value = s.value[s.cursor:]
		s.cursor = 0
	case "ctrl+k":
		s.value = s.value[:s.cursor]
	default:
		runes := []rune(msg.String())
		if len(runes) == 1 && unicode.IsPrint(runes[0]) {
			s.value = append(s.value[:s.cursor:s.cursor], append([]rune{runes[0]}, s.value[s.cursor:]...)...)
			s.cursor++
		}
	}
	return s, nil
}
