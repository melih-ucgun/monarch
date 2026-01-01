package core

import (
	"io"
	"io/fs"
	"os"
)

// FileSystem is an interface for filesystem operations
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	Lstat(name string) (fs.FileInfo, error)
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Remove(name string) error
	RemoveAll(path string) error
	Readlink(name string) (string, error)
	Symlink(oldname, newname string) error
	Chmod(name string, mode os.FileMode) error
	Open(name string) (File, error)
	Create(name string) (File, error)
	ReadDir(name string) ([]fs.DirEntry, error)
}

// File is a minimal interface for a file object
type File interface {
	io.Reader
	io.ReaderAt
	io.Writer
	io.Closer
	Stat() (fs.FileInfo, error)
}

// RealFS is a real filesystem implementation using os package
type RealFS struct{}

func (f *RealFS) Stat(name string) (fs.FileInfo, error)  { return os.Stat(name) }
func (f *RealFS) Lstat(name string) (fs.FileInfo, error) { return os.Lstat(name) }
func (f *RealFS) ReadFile(name string) ([]byte, error)   { return os.ReadFile(name) }
func (f *RealFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (f *RealFS) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (f *RealFS) Remove(name string) error                     { return os.Remove(name) }
func (f *RealFS) RemoveAll(path string) error                  { return os.RemoveAll(path) }
func (f *RealFS) Readlink(name string) (string, error)         { return os.Readlink(name) }
func (f *RealFS) Symlink(oldname, newname string) error        { return os.Symlink(oldname, newname) }
func (f *RealFS) Chmod(name string, mode os.FileMode) error    { return os.Chmod(name, mode) }
func (f *RealFS) Open(name string) (File, error)               { return os.Open(name) }
func (f *RealFS) Create(name string) (File, error)             { return os.Create(name) }
func (f *RealFS) ReadDir(name string) ([]fs.DirEntry, error)   { return os.ReadDir(name) }

// CopyFile is a helper to copy a file using the FileSystem abstraction
func CopyFile(fs FileSystem, src, dst string, mode os.FileMode) error {
	sourceFile, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := fs.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}
	return fs.Chmod(dst, mode)
}
