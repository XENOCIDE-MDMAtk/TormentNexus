package repograph

/**
 * @file graph.go
 * @module go/internal/repograph
 *
 * WHAT: Repository graph service — builds a dependency/import graph of source code,
 *       enabling semantic code navigation, impact analysis, and code search.
 *
 * WHY: Code understanding is fundamental to AI-assisted development. Knowing which
 *      files import which, what functions are defined where, and what the dependency
 *      tree looks like enables intelligent code suggestions and refactoring.
 *
 * ADDED: v1.0.0-alpha.32
 */

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// NodeType represents the type of a graph node.
type NodeType string

const (
	NodeFile       NodeType = "file"
	NodeFunction   NodeType = "function"
	NodeTypeName   NodeType = "type"
	NodeInterface  NodeType = "interface"
	NodeImport     NodeType = "import"
	NodePackage    NodeType = "package"
)

// Node represents a node in the repository graph.
type Node struct {
	ID          string   `json:"id"`
	Type        NodeType `json:"type"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Package     string   `json:"package,omitempty"`
	Language    string   `json:"language,omitempty"`
	LineStart   int      `json:"lineStart,omitempty"`
	LineEnd     int      `json:"lineEnd,omitempty"`
	IsExported  bool     `json:"isExported"`
	DocComment  string   `json:"docComment,omitempty"`
}

// Edge represents a dependency between two nodes.
type Edge struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Type    string `json:"type"` // "imports", "calls", "implements", "extends", "references"
	Weight  int    `json:"weight"`
}

// Graph is the complete repository dependency graph.
type Graph struct {
	Nodes    map[string]*Node `json:"nodes"`
	Edges    []Edge           `json:"edges"`
	RootPath string           `json:"rootPath"`
	BuiltAt  time.Time        `json:"builtAt"`
	Stats    GraphStats       `json:"stats"`
}

// GraphStats contains summary statistics about the graph.
type GraphStats struct {
	TotalFiles     int `json:"totalFiles"`
	TotalFunctions int `json:"totalFunctions"`
	TotalTypes     int `json:"totalTypes"`
	TotalImports   int `json:"totalImports"`
	TotalEdges     int `json:"totalEdges"`
}

// RepoGraphService builds and queries repository graphs.
type RepoGraphService struct {
	root string
	mu   sync.RWMutex
	graph *Graph
}

// NewRepoGraphService creates a new graph service for the given root.
func NewRepoGraphService(root string) *RepoGraphService {
	return &RepoGraphService{
		root: root,
	}
}

// Build scans the repository and builds the dependency graph.
func (rgs *RepoGraphService) Build(ctx context.Context) (*Graph, error) {
	graph := &Graph{
		Nodes:    make(map[string]*Node),
		RootPath: rgs.root,
		BuiltAt:  time.Now().UTC(),
	}

	// Walk the repo and index files
	err := filepath.WalkDir(rgs.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip common non-source directories
		if d.IsDir() {
			name := strings.ToLower(d.Name())
			switch name {
			case "node_modules", ".git", "dist", "build", "coverage", "vendor",
				"__pycache__", ".next", ".cache", "target", "bin":
				return filepath.SkipDir
			}
			return nil
		}

		// Only process source files
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".rs":
			rgs.indexFile(graph, path)
		default:
			return nil
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	rgs.mu.Lock()
	rgs.graph = graph
	rgs.mu.Unlock()

	// Calculate stats
	graph.Stats = GraphStats{
		TotalFiles:     rgs.countByType(graph, NodeFile),
		TotalFunctions: rgs.countByType(graph, NodeFunction),
		TotalTypes:     rgs.countByType(graph, NodeTypeName),
		TotalImports:   rgs.countByType(graph, NodeImport),
		TotalEdges:     len(graph.Edges),
	}

	return graph, nil
}

// GetGraph returns the current graph, building it if necessary.
func (rgs *RepoGraphService) GetGraph() *Graph {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()
	return rgs.graph
}

// FindReferences finds all nodes that reference the given symbol.
func (rgs *RepoGraphService) FindReferences(symbolName string) []*Node {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()

	if rgs.graph == nil {
		return nil
	}

	var refs []*Node
	for _, edge := range rgs.graph.Edges {
		from := rgs.graph.Nodes[edge.From]
		if edge.Type == "references" || edge.Type == "calls" {
			if to := rgs.graph.Nodes[edge.To]; to != nil && to.Name == symbolName {
				refs = append(refs, from)
			}
		}
	}
	return refs
}

// FindDependents returns all files that depend on the given file.
func (rgs *RepoGraphService) FindDependents(filePath string) []string {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()

	if rgs.graph == nil {
		return nil
	}

	fileID := "file:" + filePath
	var dependents []string
	seen := make(map[string]bool)

	var walk func(id string)
	walk = func(id string) {
		for _, edge := range rgs.graph.Edges {
			if edge.To == id && !seen[edge.From] {
				seen[edge.From] = true
				if node := rgs.graph.Nodes[edge.From]; node != nil && node.Type == NodeFile {
					dependents = append(dependents, node.Path)
				}
				walk(edge.From)
			}
		}
	}

	walk(fileID)
	return dependents
}

// SearchSymbols finds symbols matching a query.
func (rgs *RepoGraphService) SearchSymbols(query string, limit int) []*Node {
	rgs.mu.RLock()
	defer rgs.mu.RUnlock()

	if rgs.graph == nil {
		return nil
	}

	q := strings.ToLower(query)
	var results []*Node

	for _, node := range rgs.graph.Nodes {
		if node.Type == NodeFile || node.Type == NodeImport {
			continue
		}
		if strings.Contains(strings.ToLower(node.Name), q) {
			results = append(results, node)
		}
		if len(results) >= limit {
			break
		}
	}

	return results
}

// --- File Indexing ---

var (
	goFuncRe    = regexp.MustCompile(`^func\s+(?:\([^)]+\)\s+)?(\w+)`)
	goTypeRe    = regexp.MustCompile(`^type\s+(\w+)\s+(struct|interface)`)
	goImportRe  = regexp.MustCompile(`^\s*"([^"]+)"`)
	goPackageRe = regexp.MustCompile(`^package\s+(\w+)`)

	tsFuncRe    = regexp.MustCompile(`(?:export\s+)?(?:async\s+)?function\s+(\w+)`)
	tsClassRe   = regexp.MustCompile(`(?:export\s+)?(?:abstract\s+)?class\s+(\w+)`)
	tsInterfaceRe = regexp.MustCompile(`(?:export\s+)?interface\s+(\w+)`)
	tsImportRe  = regexp.MustCompile(`import.*from\s+['"]([^'"]+)['"]`)

	pyFuncRe    = regexp.MustCompile(`^def\s+(\w+)`)
	pyClassRe   = regexp.MustCompile(`^class\s+(\w+)`)
	pyImportRe  = regexp.MustCompile(`^import\s+(\S+)|^from\s+(\S+)\s+import`)

	rsFuncRe    = regexp.MustCompile(`(?:pub\s+)?fn\s+(\w+)`)
	rsStructRe  = regexp.MustCompile(`(?:pub\s+)?struct\s+(\w+)`)

	tsAbsoluteImportRe = regexp.MustCompile(`^@/|^\w`)
)

func (rgs *RepoGraphService) indexFile(graph *Graph, filePath string) {
	relPath, _ := filepath.Rel(rgs.root, filePath)
	relPath = filepath.ToSlash(relPath)

	// Create file node
	fileID := "file:" + relPath
	fileNode := &Node{
		ID:       fileID,
		Type:     NodeFile,
		Name:     filepath.Base(filePath),
		Path:     relPath,
		Language: languageFromExt(filepath.Ext(filePath)),
	}
	graph.Nodes[fileID] = fileNode

	// Read and parse
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}
	if len(data) > 500*1024 {
		return // Skip very large files
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		rgs.indexGoFile(graph, relPath, string(data))
	case ".ts", ".tsx", ".js", ".jsx":
		rgs.indexTSFile(graph, relPath, string(data))
	case ".py":
		rgs.indexPythonFile(graph, relPath, string(data))
	case ".rs":
		rgs.indexRustFile(graph, relPath, string(data))
	}
}

func (rgs *RepoGraphService) resolveTSImport(currentFile, importPath string) string {
	if tsAbsoluteImportRe.MatchString(importPath) {
		return "import:" + importPath
	}
	
	dir := filepath.Dir(currentFile)
	target := filepath.ToSlash(filepath.Clean(filepath.Join(dir, importPath)))
	
	// Try extensions
	extensions := []string{".ts", ".tsx", ".js", ".jsx", "/index.ts", "/index.tsx"}
	for _, ext := range extensions {
		fullPath := target + ext
		if _, ok := rgs.graph.Nodes["file:"+fullPath]; ok {
			return "file:" + fullPath
		}
	}
	
	return "import:" + importPath
}

func (rgs *RepoGraphService) indexGoFile(graph *Graph, relPath, content string) {
	fileID := "file:" + relPath
	pkgName := ""

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	inImportBlock := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Package
		if matches := goPackageRe.FindStringSubmatch(line); len(matches) > 1 {
			pkgName = matches[1]
			if file := graph.Nodes[fileID]; file != nil {
				file.Package = pkgName
			}
		}

		// Imports
		if strings.HasPrefix(strings.TrimSpace(line), "import (") {
			inImportBlock = true
			continue
		}
		if inImportBlock {
			if strings.TrimSpace(line) == ")" {
				inImportBlock = false
				continue
			}
			if matches := goImportRe.FindStringSubmatch(line); len(matches) > 1 {
				graph.Edges = append(graph.Edges, Edge{
					From: fileID, To: "import:" + matches[1], Type: "imports",
				})
			}
			continue
		}
		if matches := regexp.MustCompile(`import\s+"([^"]+)"`).FindStringSubmatch(line); len(matches) > 1 {
			graph.Edges = append(graph.Edges, Edge{
				From: fileID, To: "import:" + matches[1], Type: "imports",
			})
		}

		// Functions
		if matches := goFuncRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			exported := name[0] >= 'A' && name[0] <= 'Z'
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeFunction,
				Name:       name,
				Path:       relPath,
				Package:    pkgName,
				LineStart:  lineNum,
				IsExported: exported,
				Language:   "go",
			}
		}

		// Types
		if matches := goTypeRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			nodeType := NodeTypeName
 if matches[2] == "interface" {
				nodeType = NodeInterface
			}
			exported := name[0] >= 'A' && name[0] <= 'Z'
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       nodeType,
				Name:       name,
				Path:       relPath,
				Package:    pkgName,
				LineStart:  lineNum,
				IsExported: exported,
				Language:   "go",
			}
		}
	}
}

func (rgs *RepoGraphService) indexTSFile(graph *Graph, relPath, content string) {
	fileID := "file:" + relPath

	// Imports
	for _, match := range tsImportRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			target := rgs.resolveTSImport(relPath, match[1])
			graph.Edges = append(graph.Edges, Edge{
				From: fileID, To: target, Type: "imports",
			})
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matches := tsFuncRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			exported := strings.Contains(line, "export")
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeFunction,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: exported,
				Language:   "typescript",
			}
		}

		if matches := tsClassRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeTypeName,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: strings.Contains(line, "export"),
				Language:   "typescript",
			}
		}

		if matches := tsInterfaceRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeInterface,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: strings.Contains(line, "export"),
				Language:   "typescript",
			}
		}
	}
}

func (rgs *RepoGraphService) indexPythonFile(graph *Graph, relPath, content string) {
	fileID := "file:" + relPath

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matches := pyImportRe.FindStringSubmatch(line); len(matches) > 1 {
			mod := matches[1]
			if matches[2] != "" {
				mod = matches[2]
			}
			graph.Edges = append(graph.Edges, Edge{
				From: fileID, To: "import:" + mod, Type: "imports",
			})
		}

		if matches := pyFuncRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			exported := !strings.HasPrefix(name, "_")
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeFunction,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: exported,
				Language:   "python",
			}
		}

		if matches := pyClassRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeTypeName,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: !strings.HasPrefix(name, "_"),
				Language:   "python",
			}
		}
	}
}

func (rgs *RepoGraphService) indexRustFile(graph *Graph, relPath, content string) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matches := rsFuncRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			exported := strings.Contains(line, "pub")
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeFunction,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: exported,
				Language:   "rust",
			}
		}

		if matches := rsStructRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			graph.Nodes[relPath+"#"+name] = &Node{
				ID:         relPath + "#" + name,
				Type:       NodeTypeName,
				Name:       name,
				Path:       relPath,
				LineStart:  lineNum,
				IsExported: strings.Contains(line, "pub"),
				Language:   "rust",
			}
		}
	}
}

// --- Helpers ---

func languageFromExt(ext string) string {
	switch ext {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	default:
		return "unknown"
	}
}

func (rgs *RepoGraphService) countByType(graph *Graph, nodeType NodeType) int {
	count := 0
	for _, node := range graph.Nodes {
		if node.Type == nodeType {
			count++
		}
	}
	return count
}

// Ensure sort import is available
var _ = sort.Ints
