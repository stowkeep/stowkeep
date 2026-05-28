package compose

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"

	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
)

const (
	// MaxFileSize is the maximum compose file size (1 MiB).
	MaxFileSize = 1 << 20
	// MaxDepth is the maximum YAML nesting depth.
	MaxDepth = 64
	// MaxAnchors is the maximum number of YAML anchors.
	MaxAnchors = 32
)

var stackNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,62}$`)

// ValidationError describes a field-level compose validation failure.
type ValidationError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// Result is the outcome of compose validation.
type Result struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
	Hash   string            `json:"hash,omitempty"`
}

// Validate parses and validates compose YAML against the Compose spec.
func Validate(ctx context.Context, content []byte, stackName string) Result {
	if err := validateStackName(stackName); err != nil {
		return Result{Valid: false, Errors: []ValidationError{{Path: "name", Message: err.Error()}}}
	}
	if len(content) == 0 {
		return Result{Valid: false, Errors: []ValidationError{{Path: "compose", Message: "compose content is required"}}}
	}
	if len(content) > MaxFileSize {
		return Result{Valid: false, Errors: []ValidationError{{Path: "compose", Message: fmt.Sprintf("compose file exceeds %d byte limit", MaxFileSize)}}}
	}
	if err := checkYAMLDepthAndAnchors(content); err != nil {
		return Result{Valid: false, Errors: []ValidationError{{Path: "compose", Message: err.Error()}}}
	}

	project, err := loadProject(ctx, content, stackName)
	if err != nil {
		return Result{Valid: false, Errors: formatLoadError(err)}
	}
	if len(project.Services) == 0 {
		return Result{Valid: false, Errors: []ValidationError{{Path: "services", Message: "at least one service is required"}}}
	}

	return Result{Valid: true, Hash: ContentHash(content)}
}

// LoadProject parses compose YAML into a project model.
func LoadProject(ctx context.Context, content []byte, stackName string) (*types.Project, error) {
	if err := validateStackName(stackName); err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return nil, errors.New("compose content is required")
	}
	if len(content) > MaxFileSize {
		return nil, fmt.Errorf("compose file exceeds %d byte limit", MaxFileSize)
	}
	if err := checkYAMLDepthAndAnchors(content); err != nil {
		return nil, err
	}
	return loadProject(ctx, content, stackName)
}

// ContentHash returns sha256 hex of compose bytes for audit and TOCTOU checks.
func ContentHash(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// ValidateStackName checks stack name is DNS-safe for Swarm.
func ValidateStackName(name string) error {
	return validateStackName(name)
}

func validateStackName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("stack name is required")
	}
	if len(name) > 63 {
		return errors.New("stack name must be at most 63 characters")
	}
	if !stackNamePattern.MatchString(name) {
		return errors.New("stack name must start with a lowercase letter and contain only lowercase letters, digits, hyphens, and underscores")
	}
	return nil
}

func loadProject(ctx context.Context, content []byte, stackName string) (*types.Project, error) {
	cfg := types.ConfigDetails{
		ConfigFiles: []types.ConfigFile{{
			Filename: "compose.yml",
			Content:  content,
		}},
		WorkingDir: "/",
		Environment: map[string]string{
			"COMPOSE_PROJECT_NAME": stackName,
		},
	}
	return loader.LoadWithContext(ctx, cfg, func(o *loader.Options) {
		o.SetProjectName(stackName, true)
	})
}

func formatLoadError(err error) []ValidationError {
	if err == nil {
		return nil
	}
	msg := err.Error()
	path := "compose"
	if idx := strings.Index(msg, " in "); idx > 0 {
		path = strings.TrimSpace(msg[:idx])
	}
	return []ValidationError{{Path: path, Message: msg}}
}

func checkYAMLDepthAndAnchors(content []byte) error {
	depth := 0
	maxDepth := 0
	anchorCount := 0
	r := bytes.NewReader(content)
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read compose: %w", err)
		}
		switch b {
		case ' ':
			continue
		case '\t':
			continue
		case '\n', '\r':
			depth = 0
		default:
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		if b == '&' {
			anchorCount++
		}
		if b == ' ' {
			depth++
		}
	}
	if maxDepth > MaxDepth {
		return fmt.Errorf("compose YAML nesting exceeds depth limit of %d", MaxDepth)
	}
	if anchorCount > MaxAnchors {
		return fmt.Errorf("compose YAML exceeds anchor limit of %d", MaxAnchors)
	}
	// Reject non-printable control chars except common whitespace.
	for _, ch := range string(content) {
		if ch == '\n' || ch == '\r' || ch == '\t' {
			continue
		}
		if unicode.IsControl(ch) {
			return errors.New("compose contains invalid control characters")
		}
	}
	return nil
}

// ScrubEnvironmentForLog removes environment values from a string for safe logging.
func ScrubEnvironmentForLog(s string) string {
	if !strings.Contains(s, "environment:") {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.Contains(trimmed, "=") {
			if strings.Contains(strings.ToLower(lines[max(0, i-1)]), "environment:") {
				lines[i] = "  - [redacted]"
			}
		}
	}
	return strings.Join(lines, "\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
