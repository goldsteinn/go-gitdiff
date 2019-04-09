	"io"
func TestParseGitFileHeader(t *testing.T) {
	tests := map[string]struct {
		Input  string
		Output *File
		Err    bool
	}{
		"fileContentChange": {
			Input: `diff --git a/dir/file.txt b/dir/file.txt
index 1c23fcc..40a1b33 100644
--- a/dir/file.txt
+++ b/dir/file.txt
@@ -2,3 +4,5 @@
`,
			Output: &File{
				OldName:      "dir/file.txt",
				NewName:      "dir/file.txt",
				OldMode:      os.FileMode(0100644),
				OldOIDPrefix: "1c23fcc",
				NewOIDPrefix: "40a1b33",
			},
		},
		"newFile": {
			Input: `diff --git a/dir/file.txt b/dir/file.txt
new file mode 100644
index 0000000..f5711e4
--- /dev/null
+++ b/dir/file.txt
`,
			Output: &File{
				NewName:      "dir/file.txt",
				NewMode:      os.FileMode(0100644),
				OldOIDPrefix: "0000000",
				NewOIDPrefix: "f5711e4",
				IsNew:        true,
			},
		},
		"newEmptyFile": {
			Input: `diff --git a/empty.txt b/empty.txt
new file mode 100644
index 0000000..e69de29
`,
			Output: &File{
				NewName:      "empty.txt",
				NewMode:      os.FileMode(0100644),
				OldOIDPrefix: "0000000",
				NewOIDPrefix: "e69de29",
				IsNew:        true,
			},
		},
		"deleteFile": {
			Input: `diff --git a/dir/file.txt b/dir/file.txt
deleted file mode 100644
index 44cc321..0000000
--- a/dir/file.txt
+++ /dev/null
`,
			Output: &File{
				OldName:      "dir/file.txt",
				OldMode:      os.FileMode(0100644),
				OldOIDPrefix: "44cc321",
				NewOIDPrefix: "0000000",
				IsDelete:     true,
			},
		},
		"changeMode": {
			Input: `diff --git a/file.sh b/file.sh
old mode 100644
new mode 100755
`,
			Output: &File{
				OldName: "file.sh",
				NewName: "file.sh",
				OldMode: os.FileMode(0100644),
				NewMode: os.FileMode(0100755),
			},
		},
		"rename": {
			Input: `diff --git a/foo.txt b/bar.txt
similarity index 100%
rename from foo.txt
rename to bar.txt
`,
			Output: &File{
				OldName:  "foo.txt",
				NewName:  "bar.txt",
				Score:    100,
				IsRename: true,
			},
		},
		"copy": {
			Input: `diff --git a/file.txt b/copy.txt
similarity index 100%
copy from file.txt
copy to copy.txt
`,
			Output: &File{
				OldName: "file.txt",
				NewName: "copy.txt",
				Score:   100,
				IsCopy:  true,
			},
		},
		"missingDefaultFilename": {
			Input: `diff --git a/foo.sh b/bar.sh
old mode 100644
new mode 100755
`,
			Err: true,
		},
		"missingNewFilename": {
			Input: `diff --git a/file.txt b/file.txt
index 1c23fcc..40a1b33 100644
--- a/file.txt
`,
			Err: true,
		},
		"missingOldFilename": {
			Input: `diff --git a/file.txt b/file.txt
index 1c23fcc..40a1b33 100644
+++ b/file.txt
`,
			Err: true,
		},
		"invalidHeaderLine": {
			Input: `diff --git a/file.txt b/file.txt
index deadbeef
--- a/file.txt
+++ b/file.txt
`,
			Err: true,
		},
		"notGitHeader": {
			Input: `--- file.txt
+++ file.txt
@@ -0,0 +1 @@
`,
			Output: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newTestParser(test.Input, true)

			f, err := p.ParseGitFileHeader()
			if test.Err {
				if err == nil || err == io.EOF {
					t.Fatalf("expected error parsing git file header, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error parsing git file header: %v", err)
			}

			if !reflect.DeepEqual(test.Output, f) {
				t.Errorf("incorrect file\nexpected: %+v\n  actual: %+v", test.Output, f)
			}
		})
	}
}

func TestParseTraditionalFileHeader(t *testing.T) {
	tests := map[string]struct {
		Input  string
		Output *File
		Err    bool
	}{
		"fileContentChange": {
			Input: `--- dir/file_old.txt	2019-03-21 23:00:00.0 -0700
+++ dir/file_new.txt	2019-03-21 23:30:00.0 -0700
@@ -0,0 +1 @@
`,
			Output: &File{
				OldName: "dir/file_new.txt",
				NewName: "dir/file_new.txt",
			},
		},
		"newFile": {
			Input: `--- /dev/null	1969-12-31 17:00:00.0 -0700
+++ dir/file.txt	2019-03-21 23:30:00.0 -0700
@@ -0,0 +1 @@
`,
			Output: &File{
				NewName: "dir/file.txt",
				IsNew:   true,
			},
		},
		"newFileTimestamp": {
			Input: `--- dir/file.txt	1969-12-31 17:00:00.0 -0700
+++ dir/file.txt	2019-03-21 23:30:00.0 -0700
@@ -0,0 +1 @@
`,
			Output: &File{
				NewName: "dir/file.txt",
				IsNew:   true,
			},
		},
		"deleteFile": {
			Input: `--- dir/file.txt	2019-03-21 23:30:00.0 -0700
+++ /dev/null	1969-12-31 17:00:00.0 -0700
@@ -0,0 +1 @@
`,
			Output: &File{
				OldName:  "dir/file.txt",
				IsDelete: true,
			},
		},
		"deleteFileTimestamp": {
			Input: `--- dir/file.txt	2019-03-21 23:30:00.0 -0700
+++ dir/file.txt	1969-12-31 17:00:00.0 -0700
@@ -0,0 +1 @@
`,
			Output: &File{
				OldName:  "dir/file.txt",
				IsDelete: true,
			},
		},
		"useShortestPrefixName": {
			Input: `--- dir/file.txt	2019-03-21 23:00:00.0 -0700
+++ dir/file.txt~	2019-03-21 23:30:00.0 -0700
@@ -0,0 +1 @@
`,
			Output: &File{
				OldName: "dir/file.txt",
				NewName: "dir/file.txt",
			},
		},
		"notTraditionalHeader": {
			Input: `diff --git a/dir/file.txt b/dir/file.txt
--- a/dir/file.txt
+++ b/dir/file.txt
`,
			Output: nil,
		},
		"noUnifiedFragment": {
			Input: `--- dir/file_old.txt	2019-03-21 23:00:00.0 -0700
+++ dir/file_new.txt	2019-03-21 23:30:00.0 -0700
context line
+added line
`,
			Output: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newTestParser(test.Input, true)

			f, err := p.ParseTraditionalFileHeader()
			if test.Err {
				if err == nil || err == io.EOF {
					t.Fatalf("expected error parsing traditional file header, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error parsing traditional file header: %v", err)
			}

			if !reflect.DeepEqual(test.Output, f) {
				t.Errorf("incorrect file\nexpected: %+v\n  actual: %+v", test.Output, f)
			}
		})
	}
}

				if err == nil || err == io.EOF {
					t.Fatalf("expected error parsing name, but got %v", err)
				if err == nil || err == io.EOF {
					t.Fatalf("expected error parsing header data, but got %v", err)