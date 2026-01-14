package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/melih-ucgun/veto/internal/core"
	"github.com/pterm/pterm"
)

// PtermUI is an implementation of core.UI using pterm.
type PtermUI struct {
	writer io.Writer
}

// NewPtermUI creates a new PtermUI instance.
func NewPtermUI() *PtermUI {
	return &PtermUI{
		writer: os.Stdout,
	}
}

// Ensure PtermUI implements core.UI
var _ core.UI = (*PtermUI)(nil)

func (p *PtermUI) Section(title string) {
	pterm.DefaultSection.WithWriter(p.writer).Println(title)
}

func (p *PtermUI) Title(title string) {
	pterm.DefaultHeader.WithFullWidth().WithWriter(p.writer).Println(title)
}

func (p *PtermUI) Success(msg string) {
	pterm.Success.WithWriter(p.writer).Println(msg)
}

func (p *PtermUI) Info(msg string) {
	pterm.Info.WithWriter(p.writer).Println(msg)
}

func (p *PtermUI) Debug(msg string) {
	pterm.Debug.WithWriter(p.writer).Println(msg)
}

func (p *PtermUI) Warning(msg string) {
	pterm.Warning.WithWriter(p.writer).Println(msg)
}

func (p *PtermUI) Error(msg string) {
	pterm.Error.WithWriter(p.writer).Println(msg)
}

func (p *PtermUI) Printf(format string, args ...interface{}) {
	fmt.Fprintf(p.writer, format, args...)
}

func (p *PtermUI) Println(args ...interface{}) {
	fmt.Fprintln(p.writer, args...)
}

func (p *PtermUI) WithWriter(w io.Writer) core.UI {
	return &PtermUI{
		writer: w,
	}
}
