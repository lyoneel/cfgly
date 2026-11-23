package cfgly

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/lyoneel/dirly"
)

// ErrExtensionNotAllowed is returned (wrapped) by ConfigDir read and write
// operations when the file name's extension is not in the directory's
// allowed extensions. Check for it with errors.Is.
var ErrExtensionNotAllowed = errors.New("file extension not allowed")

// ConfigDir wraps dirly.Directory with config-file-specific convenience methods.
//
// A ConfigDir may restrict the files it works with to a set of extensions
// (see NewConfigDir and ConfigBuilder.WithConfigExtensions). When extensions
// are set, the filter is applied consistently to every file-content
// operation:
//
//   - Listing and search (GetByGlob, GetAllByGlobAbs/Rel, GetAllAbs/Rel) only
//     return files with an allowed extension.
//   - Exists reports false for files with any other extension.
//   - Reads and writes (ReadText, ReadTo*, WriteFrom*, BatchReadTo*,
//     BatchWriteFrom*) fail with an error wrapping ErrExtensionNotAllowed
//     before touching the file system. Batch operations validate every name
//     first, so a rejected name means nothing is read or written.
//
// Path-management operations inherited from dirly (Remove, BatchRemove,
// BatchRemoveAllFromDir, Chown, Chmod, AbsPath, SubDirectory, GetAllDirs*)
// are not filtered. Files without an extension are not allowed when a filter
// is set. With no extensions, nothing is filtered.
//
// A ConfigDir is immutable after construction and safe for concurrent use.
type ConfigDir struct {
	*dirly.Directory // Embed Directory to inherit all methods

	extensions []string // normalized: lower-case, no leading dot, unique
}

// NewConfigDir creates a ConfigDir rooted at basePath. If extensions are
// given, only files with those extensions are visible to and accepted by the
// ConfigDir (see ConfigDir). Extensions are case-insensitive and may be given
// with or without the leading dot ("yaml" and ".YAML" are equivalent); empty
// values are ignored.
func NewConfigDir(basePath string, extensions ...string) *ConfigDir {
	exts := normalizeExtensions(extensions)
	return &ConfigDir{
		Directory:  dirly.NewFilteredDirectory(basePath).WithExtensions(exts...).Build(),
		extensions: exts,
	}
}

// normalizeExtensions lower-cases extensions, strips one leading dot and
// surrounding spaces, and drops empty and duplicate values.
func normalizeExtensions(extensions []string) []string {
	var out []string
	seen := make(map[string]bool, len(extensions))
	for _, e := range extensions {
		e = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(e), "."))
		if e == "" || seen[e] {
			continue
		}
		seen[e] = true
		out = append(out, e)
	}
	return out
}

// Extensions returns a copy of the allowed extensions (lower-case, without
// leading dot). It returns nil when the directory is not filtered.
func (d *ConfigDir) Extensions() []string {
	if len(d.extensions) == 0 {
		return nil
	}
	out := make([]string, len(d.extensions))
	copy(out, d.extensions)
	return out
}

// checkExtension returns an error wrapping ErrExtensionNotAllowed when a
// filter is set and filename's extension is not allowed. Empty names are
// left to dirly so its own validation error is returned unchanged.
func (d *ConfigDir) checkExtension(filename string) error {
	if len(d.extensions) == 0 || filename == "" {
		return nil
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	for _, allowed := range d.extensions {
		if ext == allowed {
			return nil
		}
	}
	return fmt.Errorf("%w: %q (allowed: %s)", ErrExtensionNotAllowed, filename, strings.Join(d.extensions, ", "))
}

func (d *ConfigDir) checkAll(filenames []string) error {
	for _, f := range filenames {
		if err := d.checkExtension(f); err != nil {
			return err
		}
	}
	return nil
}

func checkKeys[V any](d *ConfigDir, files map[string]V) error {
	for f := range files {
		if err := d.checkExtension(f); err != nil {
			return err
		}
	}
	return nil
}

// GetByGlob Searches for files matching a pattern and returns absolute paths.
// Only files with an allowed extension are returned.
func (d *ConfigDir) GetByGlob(pattern string) ([]string, error) {
	return d.Directory.GetAllByGlobAbs(pattern)
}

// ReadText Reads a file and returns its content as a string.
func (d *ConfigDir) ReadText(filename string) (string, error) {
	return d.ReadToString(filename)
}

// Exists Checks if a file exists in the directory and has an allowed extension.
func (d *ConfigDir) Exists(filename string) bool {
	if d.checkExtension(filename) != nil {
		return false
	}
	return d.Directory.Exists(filename)
}

// ReadToBytes reads a file with an allowed extension and returns raw bytes.
func (d *ConfigDir) ReadToBytes(filename string) ([]byte, error) {
	if err := d.checkExtension(filename); err != nil {
		return nil, err
	}
	return d.Directory.ReadToBytes(filename)
}

// ReadToString reads a file with an allowed extension and returns its content as a string.
func (d *ConfigDir) ReadToString(filename string) (string, error) {
	if err := d.checkExtension(filename); err != nil {
		return "", err
	}
	return d.Directory.ReadToString(filename)
}

// ReadToLines reads a file with an allowed extension and returns its lines.
func (d *ConfigDir) ReadToLines(filename string) ([]string, error) {
	if err := d.checkExtension(filename); err != nil {
		return nil, err
	}
	return d.Directory.ReadToLines(filename)
}

// ReadToBuff opens a file with an allowed extension for buffered reading.
func (d *ConfigDir) ReadToBuff(filename string) (*os.File, *bufio.Reader, error) {
	if err := d.checkExtension(filename); err != nil {
		return nil, nil, err
	}
	return d.Directory.ReadToBuff(filename)
}

// ReadToScanner opens a file with an allowed extension for line scanning.
func (d *ConfigDir) ReadToScanner(filename string) (*os.File, *bufio.Scanner, error) {
	if err := d.checkExtension(filename); err != nil {
		return nil, nil, err
	}
	return d.Directory.ReadToScanner(filename)
}

// WriteFromBytes writes data to a file with an allowed extension.
func (d *ConfigDir) WriteFromBytes(filename string, data []byte, perm os.FileMode) error {
	if err := d.checkExtension(filename); err != nil {
		return err
	}
	return d.Directory.WriteFromBytes(filename, data, perm)
}

// WriteFromString writes s to a file with an allowed extension.
func (d *ConfigDir) WriteFromString(filename string, s string, perm os.FileMode) error {
	if err := d.checkExtension(filename); err != nil {
		return err
	}
	return d.Directory.WriteFromString(filename, s, perm)
}

// WriteFromLines writes lines to a file with an allowed extension.
func (d *ConfigDir) WriteFromLines(filename string, lines []string, perm os.FileMode) error {
	if err := d.checkExtension(filename); err != nil {
		return err
	}
	return d.Directory.WriteFromLines(filename, lines, perm)
}

// WriteFromBuff writes the content of r to a file with an allowed extension.
func (d *ConfigDir) WriteFromBuff(filename string, r *bufio.Reader, perm os.FileMode) error {
	if err := d.checkExtension(filename); err != nil {
		return err
	}
	return d.Directory.WriteFromBuff(filename, r, perm)
}

// WriteFromScanner writes the lines of s to a file with an allowed extension.
func (d *ConfigDir) WriteFromScanner(filename string, s *bufio.Scanner, perm os.FileMode) error {
	if err := d.checkExtension(filename); err != nil {
		return err
	}
	return d.Directory.WriteFromScanner(filename, s, perm)
}

// BatchReadToBytes reads several files; every name must have an allowed extension.
func (d *ConfigDir) BatchReadToBytes(filenames []string) (map[string][]byte, error) {
	if err := d.checkAll(filenames); err != nil {
		return nil, err
	}
	return d.Directory.BatchReadToBytes(filenames)
}

// BatchReadToString reads several files; every name must have an allowed extension.
func (d *ConfigDir) BatchReadToString(filenames []string) (map[string]string, error) {
	if err := d.checkAll(filenames); err != nil {
		return nil, err
	}
	return d.Directory.BatchReadToString(filenames)
}

// BatchReadToLines reads several files; every name must have an allowed extension.
func (d *ConfigDir) BatchReadToLines(filenames []string) (map[string][]string, error) {
	if err := d.checkAll(filenames); err != nil {
		return nil, err
	}
	return d.Directory.BatchReadToLines(filenames)
}

// BatchReadToBuff opens several files; every name must have an allowed extension.
func (d *ConfigDir) BatchReadToBuff(filenames []string) (map[string]*bufio.Reader, error) {
	if err := d.checkAll(filenames); err != nil {
		return nil, err
	}
	return d.Directory.BatchReadToBuff(filenames)
}

// BatchReadToScanner opens several files; every name must have an allowed extension.
func (d *ConfigDir) BatchReadToScanner(filenames []string) (map[string]*bufio.Scanner, error) {
	if err := d.checkAll(filenames); err != nil {
		return nil, err
	}
	return d.Directory.BatchReadToScanner(filenames)
}

// BatchWriteFromBytes writes several files; every name must have an allowed extension.
func (d *ConfigDir) BatchWriteFromBytes(files map[string][]byte) error {
	if err := checkKeys(d, files); err != nil {
		return err
	}
	return d.Directory.BatchWriteFromBytes(files)
}

// BatchWriteFromString writes several files; every name must have an allowed extension.
func (d *ConfigDir) BatchWriteFromString(files map[string]string) error {
	if err := checkKeys(d, files); err != nil {
		return err
	}
	return d.Directory.BatchWriteFromString(files)
}

// BatchWriteFromLines writes several files; every name must have an allowed extension.
func (d *ConfigDir) BatchWriteFromLines(files map[string][]string) error {
	if err := checkKeys(d, files); err != nil {
		return err
	}
	return d.Directory.BatchWriteFromLines(files)
}

// BatchWriteFromBuff writes several files; every name must have an allowed extension.
func (d *ConfigDir) BatchWriteFromBuff(files map[string]*bufio.Reader) error {
	if err := checkKeys(d, files); err != nil {
		return err
	}
	return d.Directory.BatchWriteFromBuff(files)
}

// BatchWriteFromScanner writes several files; every name must have an allowed extension.
func (d *ConfigDir) BatchWriteFromScanner(files map[string]*bufio.Scanner) error {
	if err := checkKeys(d, files); err != nil {
		return err
	}
	return d.Directory.BatchWriteFromScanner(files)
}
