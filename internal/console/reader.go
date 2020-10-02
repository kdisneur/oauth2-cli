package console

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

// Reader is in charge of reading console input
type Reader struct {
	b *bufio.Reader
}

// ReadStringOptions represents all possible options when reading a string
type ReadStringOptions struct {
	DefaultValue string
	Required     bool
	Writer       io.Writer
	Prompt       string
}

// NewReader initialize a new console input reader
func NewReader(r io.Reader) *Reader {
	return &Reader{b: bufio.NewReader(r)}
}

// ReadStringlnWithOptions reads a string until newline happen.
// If string is empty returns the default value
func (r *Reader) ReadStringlnWithOptions(opts ReadStringOptions) (string, error) {
	w := ioutil.Discard
	if opts.Writer != nil {
		w = opts.Writer
	}

	prompt := fmt.Sprintf("%s: ", opts.Prompt)
	if opts.DefaultValue != "" {
		prompt = fmt.Sprintf("%s (%s): ", opts.Prompt, opts.DefaultValue)
	}

	fmt.Fprintf(w, prompt)

	value, err := r.ReadStringln()
	if err != nil {
		return opts.DefaultValue, err
	}

	if opts.Required && value == "" && opts.DefaultValue == "" {
		return r.ReadStringlnWithOptions(opts)
	}

	if value == "" {
		return opts.DefaultValue, nil
	}

	return value, nil
}

// ReadStringln reads a string until newline happen
func (r *Reader) ReadStringln() (string, error) {
	value, err := r.b.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("can't read string: %v", err)
	}

	return strings.TrimSpace(value), err
}
