package commonlog

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const timeLayout = "02/Jan/2006:15:04:05 -0700"

// Regex to extract the section name
var sectionRegex = regexp.MustCompile(`\/([0-9a-zA-Z+.-]+)[\/&| ]`)

// Lexer to read a common log format
type lexer struct {
	position int
	line     string
}

// Parse parses the line and returns a log Event
func Parse(line string) (event Event, err error) {
	if len(line) == 0 {
		return event, errors.New("empty log line")
	}

	l := lexer{
		line: line,
	}

	// host
	value, err := l.nextField(' ')
	if err != nil {
		return event, fmt.Errorf("reading host: %w - event: %s", err, event)
	}
	event.Host = value

	// RFC931
	value, err = l.nextField(' ')
	if err != nil {
		return event, fmt.Errorf("reading rfc931: %w - event: %s", err, event)
	}
	event.RFC931 = value

	// User
	value, err = l.nextField(' ')
	if err != nil {
		return event, fmt.Errorf("reading user: %w - event: %s", err, event)
	}
	event.User = value

	// Date
	err = l.except('[')
	if err != nil {
		return event, fmt.Errorf("reading date: %w - event: %s", err, event)
	}

	value, err = l.nextField(']')
	if err != nil {
		return event, fmt.Errorf("reading date: %w - event: %s", err, event)
	}
	event.Date, err = time.Parse(timeLayout, value)
	if err != nil {
		return event, fmt.Errorf("invalid date format: %w", err)
	}

	err = l.except(' ')
	if err != nil {
		return event, fmt.Errorf("reading request: %w - event: %s", err, event)
	}

	// Request
	err = l.except('"')
	if err != nil {
		return event, fmt.Errorf("reading request: %w - event: %s", err, event)
	}

	value, err = l.nextField('"')
	if err != nil {
		return event, fmt.Errorf("reading request: %w - event: %s", err, event)
	}
	event.Request = value
	section := sectionRegex.FindString(value)
	if len(section) == 0 {
		return event, fmt.Errorf("section not found: request %s", event.Request)
	}
	event.Section = section[1 : len(section)-1] // remove the first / and remove the last character which can a / or a space

	err = l.except(' ')
	if err != nil {
		return event, fmt.Errorf("reading status: %w - event: %s", err, event)
	}

	// Status code
	value, err = l.nextField(' ')
	if err != nil {
		return event, fmt.Errorf("reading status: %w - event: %s", err, event)
	}
	event.Status, err = strconv.Atoi(value)
	if err != nil {
		return event, fmt.Errorf("invalid status format: %w", err)
	}

	// Bytes
	value, err = l.nextField(' ')
	if err != nil {
		// last field
		value = l.line[l.position:]
	}
	event.Bytes, err = strconv.Atoi(value)
	if err != nil {
		return event, fmt.Errorf("invalid bytes number: %w", err)
	}
	return event, nil
}

func (l *lexer) nextField(separator byte) (string, error) {
	var buffer strings.Builder
	for i := l.position; i < len(l.line); i++ {
		if l.line[i] == separator {
			l.position = i + 1
			return buffer.String(), nil
		}
		buffer.WriteByte(l.line[i])
	}

	return "", fmt.Errorf("separator not found %c", separator)
}

func (l *lexer) except(rune byte) error {
	if l.line[l.position] == rune {
		l.position++
		return nil
	}

	return fmt.Errorf("character not found %c", rune)
}
