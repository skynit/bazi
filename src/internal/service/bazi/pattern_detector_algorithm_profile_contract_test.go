package bazi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type patternAlgorithmASTEntry struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

func loadPatternProductionFunctions(t *testing.T) (*token.FileSet, map[string]*ast.FuncDecl) {
	t.Helper()
	fset := token.NewFileSet()
	functions := make(map[string]*ast.FuncDecl)
	parsed, err := parser.ParseDir(fset, ".", func(info fs.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	pkg := parsed["bazi"]
	if pkg == nil {
		t.Fatal("production bazi package not found")
	}
	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil {
				functions[function.Name.Name] = function
			}
		}
	}
	return fset, functions
}

func patternAlgorithmCallClosure(t *testing.T, root string, functions map[string]*ast.FuncDecl) []string {
	t.Helper()
	if functions[root] == nil {
		t.Fatalf("algorithm root function %q not found", root)
	}
	seen := make(map[string]struct{})
	boundaries := map[string]struct{}{"fixedPatternDetection": {}}
	queue := []string{root}
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		ast.Inspect(functions[name].Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			identifier, ok := call.Fun.(*ast.Ident)
			if !ok {
				return true
			}
			_, boundary := boundaries[identifier.Name]
			if !boundary && functions[identifier.Name] != nil {
				queue = append(queue, identifier.Name)
			}
			return true
		})
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func patternAlgorithmASTSHA256(t *testing.T, fset *token.FileSet, functions map[string]*ast.FuncDecl, names []string) string {
	t.Helper()
	entries := make([]patternAlgorithmASTEntry, 0, len(names))
	for _, name := range names {
		var source bytes.Buffer
		if err := format.Node(&source, fset, functions[name]); err != nil {
			t.Fatal(err)
		}
		entries = append(entries, patternAlgorithmASTEntry{Name: name, Source: source.String()})
	}
	payload, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func TestPatternDetectorAlgorithmProfilesBindProductionCallClosures(t *testing.T) {
	fset, functions := loadPatternProductionFunctions(t)
	for _, detector := range patternDetectorRegistry() {
		profile, ok := patternDetectorAlgorithmProfileForRule(detector.ruleID)
		if !ok {
			t.Errorf("detector %s has no algorithm Profile", detector.ruleID)
			continue
		}
		closure := patternAlgorithmCallClosure(t, profile.RootFunction, functions)
		digest := patternAlgorithmASTSHA256(t, fset, functions, closure)
		if profile.Scheme != "go_ast_detector_closure_v1" || !reflect.DeepEqual(profile.Functions, closure) || profile.ASTSHA256 != digest {
			t.Errorf("detector %s algorithm Profile:\n functions: %#v\n sha256: %s", detector.ruleID, closure, digest)
		}
		if detector.algorithmSHA256 != profile.ASTSHA256 {
			t.Errorf("detector %s registry algorithm SHA-256 = %q, want %q", detector.ruleID, detector.algorithmSHA256, profile.ASTSHA256)
		}
	}
	if _, ok := patternDetectorAlgorithmProfileForRule("pattern.unknown"); ok {
		t.Fatal("unknown detector received an algorithm Profile")
	}
}

func TestPatternDetectorAlgorithmProfileMetadataContract(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		if count := strings.Count(table.Description, "旧逐规则Profile仍以人工implementation字符串代表算法实现"); count != 1 {
			t.Errorf("algorithm Profile metadata statement count = %d, want 1", count)
		}
		for _, fragment := range []string{
			"pattern-candidate-set-v25删除注册表人工implementation字段",
			"go_ast_detector_closure_v1算法Profile",
			"规范排序的同包调用闭包与规范化Go AST SHA-256",
			"fixedPatternDetection作为已由输出名Profile独立约束的边界",
			"构建期合同从生产源码重算闭包和摘要，运行时不读取源码",
			"bbb80d8b291e81264f6894933b422c45ad40940138d49e58bcd8ad8b98d1048f",
			"候选命中、封闭表、古籍来源和未裁决边界不变",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
