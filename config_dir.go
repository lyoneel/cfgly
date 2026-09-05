package cfgly

import (
	"gitlab.com/lyoneel/dirly"
)

// ConfigDir wraps dirly.Directory with config-file-specific convenience methods.
type ConfigDir struct {
	*dirly.Directory // Embed Directory to inherit all methods
}

// NewConfigDir creates a ConfigDir that will always filter specific file extensions.
func NewConfigDir(basePath string, extensions ...string) *ConfigDir {
	return &ConfigDir{
		Directory: dirly.NewFilteredDirectory(basePath).WithExtensions(extensions...).Build(),
	}
}

// GetByGlob Searches for files matching a pattern and returns absolute paths.
func (d *ConfigDir) GetByGlob(pattern string) ([]string, error) {
	return d.Directory.GetAllByGlobAbs(pattern)
}

// ReadText Reads a file and returns its content as a string.
func (d *ConfigDir) ReadText(filename string) (string, error) {
	return d.Directory.ReadToString(filename)
}

// Exists Checks if a file exists in the directory.
func (d *ConfigDir) Exists(filename string) bool {
	return d.Directory.Exists(filename)
}
