package goldens

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"sigs.k8s.io/yaml"
)

type GoldenFile struct {
	testCase string
	testName string
	path     string
	dir      string
	filename string
}

// CheckOrUpdate writes the golden file if update flag is set
// Otherwise it checks that the contents of the file match the yaml of in
// Struct tags must be set
// usage:
// var update = flag.Bool("updategolden", false, "update golden test output files")
// err := g.CheckOrUpdate(*update, got)
func (g *GoldenFile) CheckOrUpdate(update bool, in interface{}) error {
	got, err := yaml.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshalling result as yaml: %w", err)
	}

	if update { // updating, write
		return g.Write(got)
	}
	want := g.MustRead()

	if !bytes.Equal(got, want) {
		// FAIL - calculate diff and construct error of want, got, diff
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(want), string(got), false)
		return fmt.Errorf("---WANT---\n%s\n---GOT---\n%s\n---DIFF---\n%s", want, got, dmp.DiffPrettyText(diffs))
	}
	return nil
}

func (g *GoldenFile) MustRead() []byte {
	data, err := g.Read()
	if err != nil {
		return []byte(err.Error())
	}
	return data
}

// reads the golden file
func (g *GoldenFile) Read() ([]byte, error) {
	expected, err := os.ReadFile(g.path)
	if err != nil {
		return expected, fmt.Errorf(
			"reading golden file: Test: %s Case: %s filepath: %s error:%w",
			g.testName, g.testCase, g.path, err,
		)
	}
	return expected, nil
}

// New Creates a new golden test file struct for the calling function
// Doesn't create the file itself (call Write())
func New(tb testing.TB, testCase string) (*GoldenFile, error) {
	tb.Helper()
	testName := tb.Name()
	// subtests have the form TestName/SubTestName
	testName = strings.Split(testName, "/")[0]

	g := GoldenFile{
		testName: testName,
		testCase: testCase,
		dir:      filepath.Join("testdata", testName),
		filename: testCase + ".golden",
	}

	g.path = filepath.Join(g.dir, g.filename)
	return &g, nil
}

// writes the golden file
func (g *GoldenFile) Write(data []byte) error {
	_ = os.MkdirAll(g.dir, os.ModePerm)
	err := os.WriteFile(g.path, data, 0644)
	if err != nil {
		return fmt.Errorf("writing golden file: Test: %s Case: %s filepath: %s error:%w", g.testName, g.testCase, g.path, err)
	}
	return nil
}
