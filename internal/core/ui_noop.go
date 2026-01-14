package core

import "io"

// NoOpUI is a no-operation implementation of UI.
type NoOpUI struct{}

func (n *NoOpUI) Section(title string)                      {}
func (n *NoOpUI) Title(title string)                        {}
func (n *NoOpUI) Success(msg string)                        {}
func (n *NoOpUI) Info(msg string)                           {}
func (n *NoOpUI) Debug(msg string)                          {}
func (n *NoOpUI) Warning(msg string)                        {}
func (n *NoOpUI) Error(msg string)                          {}
func (n *NoOpUI) Printf(format string, args ...interface{}) {}
func (n *NoOpUI) Println(args ...interface{})               {}
func (n *NoOpUI) WithWriter(w io.Writer) UI                 { return n }
