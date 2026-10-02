package source

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	initialWorkingDir string
	initModuleRoot    string
	initModulePath    string
)

func init() {
	initialWorkingDir, _ = os.Getwd()
	if initialWorkingDir != "" {
		initModuleRoot, initModulePath = FindModule(initialWorkingDir)
	}
}

// ResolveSourceFile resolves a source filename (which may be trimmed by -trimpath
// or modified by Bazel) to an absolute path on disk.
func ResolveSourceFile(filename string) (string, error) {
	if inBazelTest && !filepath.IsAbs(filename) {
		return bazelSourcePath(filename)
	}

	if fileExists(filename) {
		return filename, nil
	}

	modRoot := initModuleRoot
	modPath := initModulePath

	if envDir := os.Getenv("GOTESTTOOLS_MODULE_DIR"); envDir != "" {
		if root, path := FindModule(envDir); root != "" {
			modRoot = root
			modPath = path
		} else {
			modRoot = envDir
		}
	}

	if candidate := resolveInModule(modRoot, modPath, filename); candidate != "" {
		return candidate, nil
	}

	if modRoot != "" {
		candidate := filepath.Join(modRoot, filepath.FromSlash(filename))
		if fileExists(candidate) {
			return candidate, nil
		}
	}

	if initialWorkingDir != "" {
		candidate := filepath.Join(initialWorkingDir, filepath.FromSlash(filename))
		if fileExists(candidate) {
			return candidate, nil
		}
		candidate = filepath.Join(initialWorkingDir, filepath.Base(filename))
		if fileExists(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf(missingSourceMsg, filename)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func resolveInModule(modRoot, modPath, filename string) string {
	if modRoot == "" || modPath == "" {
		return ""
	}
	var rel string
	switch {
	case filename == modPath:
		rel = ""
	case strings.HasPrefix(filename, modPath+"/"):
		rel = strings.TrimPrefix(filename, modPath+"/")
	default:
		return ""
	}
	candidate := filepath.Join(modRoot, filepath.FromSlash(rel))
	if fileExists(candidate) {
		return candidate
	}
	return ""
}

var missingSourceMsg = "failed to find source file %s: " +
	"if running with -trimpath, ensure the file is within the Go module " +
	"or set GOTESTTOOLS_MODULE_DIR to the module root"

// FindModule walks up from dir looking for a go.mod file and returns the
// module root directory and the module path parsed from go.mod.
func FindModule(dir string) (string, string) {
	dir = filepath.Clean(dir)
	for {
		goMod := filepath.Join(dir, "go.mod")
		if info, err := os.Stat(goMod); err == nil && !info.IsDir() {
			if modPath, err := ParseModulePath(goMod); err == nil {
				return dir, modPath
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", ""
}

// ParseModulePath reads a go.mod file and returns the module path.
func ParseModulePath(goModPath string) (string, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inModuleBlock := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if inModuleBlock {
			if strings.HasPrefix(line, ")") {
				break
			}
			parts := strings.Fields(line)
			if len(parts) > 0 {
				return strings.Trim(parts[0], `"`), nil
			}
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 || parts[0] != "module" {
			continue
		}
		if parts[1] == "(" {
			inModuleBlock = true
			continue
		}
		return strings.Trim(parts[1], `"`), nil
	}
	return "", errors.New("module directive not found in go.mod")
}
