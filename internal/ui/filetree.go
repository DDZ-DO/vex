package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileNode represents a file or directory in the tree.
type FileNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*FileNode
	Parent   *FileNode
}

// FileTree manages a directory tree structure.
type FileTree struct {
	Root     *FileNode
	Expanded map[string]bool
}

// NewFileTree creates a new file tree.
func NewFileTree() *FileTree {
	return &FileTree{
		Expanded: make(map[string]bool),
	}
}

// LoadDirectory loads a directory into the file tree.
func (ft *FileTree) LoadDirectory(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		absPath = filepath.Dir(absPath)
	}

	ft.Root = &FileNode{
		Name:  filepath.Base(absPath),
		Path:  absPath,
		IsDir: true,
	}

	// Load children
	ft.loadChildren(ft.Root)

	// Expand root by default
	ft.Expanded[absPath] = true

	return nil
}

// loadChildren loads the children of a directory node.
func (ft *FileTree) loadChildren(node *FileNode) error {
	if !node.IsDir {
		return nil
	}

	entries, err := os.ReadDir(node.Path)
	if err != nil {
		return err
	}

	node.Children = nil

	// Separate dirs and files for sorting
	var dirs, files []*FileNode

	for _, entry := range entries {
		// Skip hidden files
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		child := &FileNode{
			Name:   entry.Name(),
			Path:   filepath.Join(node.Path, entry.Name()),
			IsDir:  entry.IsDir(),
			Parent: node,
		}

		if entry.IsDir() {
			dirs = append(dirs, child)
		} else {
			files = append(files, child)
		}
	}

	// Sort alphabetically
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	// Directories first, then files
	node.Children = append(dirs, files...)

	return nil
}

// Toggle toggles the expanded state of a directory.
func (ft *FileTree) Toggle(path string) {
	if ft.Expanded[path] {
		delete(ft.Expanded, path)
	} else {
		ft.Expanded[path] = true
		// Load children if needed
		node := ft.FindNode(path)
		if node != nil && node.IsDir && len(node.Children) == 0 {
			ft.loadChildren(node)
		}
	}
}

// IsExpanded returns whether a path is expanded.
func (ft *FileTree) IsExpanded(path string) bool {
	return ft.Expanded[path]
}

// Expand expands a directory.
func (ft *FileTree) Expand(path string) {
	ft.Expanded[path] = true
	node := ft.FindNode(path)
	if node != nil && node.IsDir && len(node.Children) == 0 {
		ft.loadChildren(node)
	}
}

// Collapse collapses a directory.
func (ft *FileTree) Collapse(path string) {
	delete(ft.Expanded, path)
}

// FindNode finds a node by its path.
func (ft *FileTree) FindNode(path string) *FileNode {
	if ft.Root == nil {
		return nil
	}
	return ft.findNodeRecursive(ft.Root, path)
}

func (ft *FileTree) findNodeRecursive(node *FileNode, path string) *FileNode {
	if node.Path == path {
		return node
	}

	for _, child := range node.Children {
		if found := ft.findNodeRecursive(child, path); found != nil {
			return found
		}
	}

	return nil
}

// GetVisibleNodes returns all visible nodes for rendering.
func (ft *FileTree) GetVisibleNodes() []*FileNode {
	if ft.Root == nil {
		return nil
	}

	var nodes []*FileNode
	ft.collectVisibleNodes(ft.Root, &nodes, 0)
	return nodes
}

func (ft *FileTree) collectVisibleNodes(node *FileNode, nodes *[]*FileNode, depth int) {
	*nodes = append(*nodes, node)

	if node.IsDir && ft.Expanded[node.Path] {
		for _, child := range node.Children {
			ft.collectVisibleNodes(child, nodes, depth+1)
		}
	}
}

// GetNodeDepth returns the depth of a node in the tree.
func (ft *FileTree) GetNodeDepth(node *FileNode) int {
	depth := 0
	current := node.Parent
	for current != nil {
		depth++
		current = current.Parent
	}
	return depth
}

// Refresh reloads the file tree.
func (ft *FileTree) Refresh() error {
	if ft.Root == nil {
		return nil
	}

	// Save expanded state
	expanded := make(map[string]bool)
	for k, v := range ft.Expanded {
		expanded[k] = v
	}

	// Reload from root
	err := ft.LoadDirectory(ft.Root.Path)
	if err != nil {
		return err
	}

	// Restore expanded state
	ft.Expanded = expanded

	return nil
}

// GetFileIcon returns an icon for the file type.
func GetFileIcon(node *FileNode) string {
	if node.IsDir {
		return "+"
	}

	ext := strings.ToLower(filepath.Ext(node.Name))
	switch ext {
	case ".go":
		return " "
	case ".py":
		return " "
	case ".js", ".jsx":
		return " "
	case ".ts", ".tsx":
		return " "
	case ".rs":
		return " "
	case ".md":
		return " "
	case ".json":
		return " "
	case ".yaml", ".yml":
		return " "
	case ".html":
		return " "
	case ".css":
		return " "
	case ".sh":
		return " "
	default:
		return " "
	}
}
