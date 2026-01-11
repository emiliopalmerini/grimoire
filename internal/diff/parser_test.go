package diff

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected int // number of files
	}{
		{
			name:     "empty diff",
			diff:     "",
			expected: 0,
		},
		{
			name: "single file single hunk",
			diff: `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}`,
			expected: 1,
		},
		{
			name: "two files",
			diff: `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"
diff --git a/util.go b/util.go
--- a/util.go
+++ b/util.go
@@ -5,2 +5,3 @@
 func helper() {
+	return
 }`,
			expected: 2,
		},
		{
			name: "new file",
			diff: `diff --git a/new.go b/new.go
new file mode 100644
--- /dev/null
+++ b/new.go
@@ -0,0 +1,5 @@
+package main
+
+func newFunc() {
+	return
+}`,
			expected: 1,
		},
		{
			name: "deleted file",
			diff: `diff --git a/old.go b/old.go
deleted file mode 100644
--- a/old.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package main
-
-func old() {}`,
			expected: 1,
		},
		{
			name: "binary file",
			diff: `diff --git a/image.png b/image.png
Binary files a/image.png and b/image.png differ`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := Parse(tt.diff)
			if len(files) != tt.expected {
				t.Errorf("Parse() got %d files, want %d", len(files), tt.expected)
			}
		})
	}
}

func TestParseHunks(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}
@@ -10,5 +11,6 @@
 func other() {
+	fmt.Println("hello")
 }`

	files := Parse(diff)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if len(files[0].Hunks) != 2 {
		t.Errorf("expected 2 hunks, got %d", len(files[0].Hunks))
	}

	// Check first hunk
	h1 := files[0].Hunks[0]
	if h1.OldStart != 1 || h1.NewStart != 1 {
		t.Errorf("hunk 1: got OldStart=%d NewStart=%d, want 1,1", h1.OldStart, h1.NewStart)
	}

	// Check second hunk
	h2 := files[0].Hunks[1]
	if h2.OldStart != 10 || h2.NewStart != 11 {
		t.Errorf("hunk 2: got OldStart=%d NewStart=%d, want 10,11", h2.OldStart, h2.NewStart)
	}
}

func TestCountLines(t *testing.T) {
	diff := `@@ -1,3 +1,4 @@
 package main

+import "fmt"
-import "os"
 func main() {}`

	count := CountLines(diff)
	if count != 2 { // 1 addition + 1 deletion
		t.Errorf("CountLines() = %d, want 2", count)
	}
}

func TestParseNewFile(t *testing.T) {
	diff := `diff --git a/new.go b/new.go
new file mode 100644
--- /dev/null
+++ b/new.go
@@ -0,0 +1,3 @@
+package main
+
+func new() {}`

	files := Parse(diff)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if !files[0].IsNew {
		t.Error("expected IsNew to be true")
	}
}

func TestParseDeletedFile(t *testing.T) {
	diff := `diff --git a/old.go b/old.go
deleted file mode 100644
--- a/old.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package main
-
-func old() {}`

	files := Parse(diff)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if !files[0].IsDelete {
		t.Error("expected IsDelete to be true")
	}
}

func TestParseBinaryFile(t *testing.T) {
	diff := `diff --git a/image.png b/image.png
Binary files a/image.png and b/image.png differ`

	files := Parse(diff)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if !files[0].IsBinary {
		t.Error("expected IsBinary to be true")
	}
}
