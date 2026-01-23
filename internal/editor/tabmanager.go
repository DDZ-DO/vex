package editor

import "path/filepath"

// TabManager manages multiple open tabs.
type TabManager struct {
	tabs      []*TabState
	activeIdx int
}

// NewTabManager creates a new tab manager with one empty tab.
func NewTabManager() *TabManager {
	tm := &TabManager{
		tabs:      []*TabState{NewTabState()},
		activeIdx: 0,
	}
	return tm
}

// ActiveTab returns the currently active tab.
func (tm *TabManager) ActiveTab() *TabState {
	if len(tm.tabs) == 0 {
		return nil
	}
	return tm.tabs[tm.activeIdx]
}

// ActiveIndex returns the index of the active tab.
func (tm *TabManager) ActiveIndex() int {
	return tm.activeIdx
}

// Tabs returns all tabs.
func (tm *TabManager) Tabs() []*TabState {
	return tm.tabs
}

// TabCount returns the number of open tabs.
func (tm *TabManager) TabCount() int {
	return len(tm.tabs)
}

// AddTab adds a new empty tab and makes it active.
func (tm *TabManager) AddTab() *TabState {
	tab := NewTabState()
	tm.tabs = append(tm.tabs, tab)
	tm.activeIdx = len(tm.tabs) - 1
	return tab
}

// AddTabFromFile adds a tab from a file. Returns existing tab if already open.
func (tm *TabManager) AddTabFromFile(path string) (*TabState, error) {
	// Normalize path for comparison
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path // Fallback to original path
	}

	// Check if file is already open
	for i, tab := range tm.tabs {
		tabPath := tab.Filepath()
		tabAbsPath, err := filepath.Abs(tabPath)
		if err != nil {
			tabAbsPath = tabPath
		}
		if tabAbsPath == absPath {
			tm.activeIdx = i
			return tab, nil
		}
	}

	// Create new tab
	tab, err := NewTabStateFromFile(path)
	if err != nil {
		return nil, err
	}

	tm.tabs = append(tm.tabs, tab)
	tm.activeIdx = len(tm.tabs) - 1
	return tab, nil
}

// CloseTab closes the tab at the given index.
// Returns true if successful, false if it was the last tab (keeps empty tab).
func (tm *TabManager) CloseTab(idx int) bool {
	if idx < 0 || idx >= len(tm.tabs) {
		return false
	}

	// If only one tab, replace with empty tab
	if len(tm.tabs) == 1 {
		tm.tabs[0] = NewTabState()
		return true
	}

	// Remove the tab
	tm.tabs = append(tm.tabs[:idx], tm.tabs[idx+1:]...)

	// Adjust active index
	if tm.activeIdx >= len(tm.tabs) {
		tm.activeIdx = len(tm.tabs) - 1
	} else if tm.activeIdx > idx {
		tm.activeIdx--
	}

	return true
}

// CloseActiveTab closes the currently active tab.
func (tm *TabManager) CloseActiveTab() bool {
	return tm.CloseTab(tm.activeIdx)
}

// SwitchTab switches to the tab at the given index.
func (tm *TabManager) SwitchTab(idx int) {
	if idx >= 0 && idx < len(tm.tabs) {
		tm.activeIdx = idx
	}
}

// NextTab switches to the next tab.
func (tm *TabManager) NextTab() {
	if len(tm.tabs) > 1 {
		tm.activeIdx = (tm.activeIdx + 1) % len(tm.tabs)
	}
}

// PrevTab switches to the previous tab.
func (tm *TabManager) PrevTab() {
	if len(tm.tabs) > 1 {
		tm.activeIdx = (tm.activeIdx - 1 + len(tm.tabs)) % len(tm.tabs)
	}
}

// IsModified returns true if any tab has unsaved changes.
func (tm *TabManager) IsModified() bool {
	for _, tab := range tm.tabs {
		if tab.Modified() {
			return true
		}
	}
	return false
}

// GetModifiedTabs returns all tabs with unsaved changes.
func (tm *TabManager) GetModifiedTabs() []*TabState {
	var modified []*TabState
	for _, tab := range tm.tabs {
		if tab.Modified() {
			modified = append(modified, tab)
		}
	}
	return modified
}

// GetModifiedPaths returns paths of all modified files.
func (tm *TabManager) GetModifiedPaths() []string {
	var paths []string
	for _, tab := range tm.tabs {
		if tab.Modified() {
			paths = append(paths, tab.Filepath())
		}
	}
	return paths
}

// SaveAll saves all modified tabs.
func (tm *TabManager) SaveAll() error {
	for _, tab := range tm.tabs {
		if tab.Modified() && tab.Filepath() != "" {
			if err := tab.Buffer().Save(); err != nil {
				return err
			}
		}
	}
	return nil
}

// FindTabByPath returns the index of the tab with the given path, or -1 if not found.
func (tm *TabManager) FindTabByPath(path string) int {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	for i, tab := range tm.tabs {
		tabPath := tab.Filepath()
		tabAbsPath, err := filepath.Abs(tabPath)
		if err != nil {
			tabAbsPath = tabPath
		}
		if tabAbsPath == absPath {
			return i
		}
	}
	return -1
}
