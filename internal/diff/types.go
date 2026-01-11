package diff

// Hunk represents a single change block in a diff.
type Hunk struct {
	FilePath string
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Content  string
	Score    float64
	Symbols  []string // Symbol names affected by this hunk
}

// FileDiff represents all changes to a single file.
type FileDiff struct {
	OldPath  string
	NewPath  string
	Hunks    []Hunk
	IsBinary bool
	IsNew    bool
	IsDelete bool
	IsRename bool
}

// DiffStats contains aggregate statistics about a diff.
type DiffStats struct {
	TotalFiles     int
	TotalLines     int
	SourceFiles    int
	SourceLines    int
	TestFiles      int
	TestLines      int
	ConfigFiles    int
	ConfigLines    int
	DocFiles       int
	DocLines       int
	GeneratedFiles int
	GeneratedLines int
}

// PrioritizedDiff is the result of diff prioritization.
type PrioritizedDiff struct {
	HighPriority string    // Full diff content for important changes
	Summary      string    // Summary of low-priority changes
	Stats        DiffStats // Aggregate statistics
}

// FileCategory represents the type of file for scoring purposes.
type FileCategory int

const (
	CategorySource FileCategory = iota
	CategoryTest
	CategoryConfig
	CategoryDoc
	CategoryGenerated
	CategoryUnknown
)

// CategoryMultiplier returns the scoring multiplier for a file category.
func (c FileCategory) Multiplier() float64 {
	switch c {
	case CategorySource:
		return 1.0
	case CategoryTest:
		return 0.6
	case CategoryConfig:
		return 0.4
	case CategoryDoc:
		return 0.2
	case CategoryGenerated:
		return 0.1
	default:
		return 0.5
	}
}

func (c FileCategory) String() string {
	switch c {
	case CategorySource:
		return "source"
	case CategoryTest:
		return "test"
	case CategoryConfig:
		return "config"
	case CategoryDoc:
		return "doc"
	case CategoryGenerated:
		return "generated"
	default:
		return "unknown"
	}
}
