package filterrepo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// TestBasicRewrite tests basic path filtering and renaming
func TestBasicRewrite(t *testing.T) {
	// Create source repository
	srcDir := t.TempDir()
	srcRepo, err := git.PlainInit(srcDir, false)
	if err != nil {
		t.Fatalf("Failed to init source repo: %v", err)
	}

	wt, err := srcRepo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	// Create initial commit with cmd/a/file.txt
	if err := os.MkdirAll(filepath.Join(srcDir, "cmd", "a"), 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "cmd", "a", "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if _, err := wt.Add("cmd/a/file.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	sig := &object.Signature{
		Name:  "Test User",
		Email: "test@example.com",
		When:  time.Now(),
	}
	_, err = wt.Commit("Initial commit", &git.CommitOptions{
		Author: sig,
	})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create second commit with cmd/a/file2.txt
	if err := os.WriteFile(filepath.Join(srcDir, "cmd", "a", "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}
	if _, err := wt.Add("cmd/a/file2.txt"); err != nil {
		t.Fatalf("Failed to add file2: %v", err)
	}
	commit2, err := wt.Commit("Second commit", &git.CommitOptions{
		Author: sig,
	})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create destination repository
	dstDir := t.TempDir()
	dstRepo, err := git.PlainInit(dstDir, true)
	if err != nil {
		t.Fatalf("Failed to init dest repo: %v", err)
	}

	// Rewrite: keep cmd/a, rename to cmd/b
	rw, err := NewRewriter(srcRepo.Storer, dstRepo.Storer, Options{
		Start:      commit2,
		KeepPrefix: "cmd/a",
		NewPrefix:  "cmd/b",
		PruneEmpty: true,
	})
	if err != nil {
		t.Fatalf("Failed to create rewriter: %v", err)
	}

	res, err := rw.Rewrite(context.Background())
	if err != nil {
		t.Fatalf("Rewrite failed: %v", err)
	}

	// Verify results
	if res.RewrittenCommits != 2 {
		t.Errorf("Expected 2 rewritten commits, got %d", res.RewrittenCommits)
	}
	if res.PrunedCommits != 0 {
		t.Errorf("Expected 0 pruned commits, got %d", res.PrunedCommits)
	}

	// Verify new tip commit exists
	newCommit, err := object.GetCommit(dstRepo.Storer, res.NewTip)
	if err != nil {
		t.Fatalf("Failed to get new tip commit: %v", err)
	}

	// Verify tree structure
	tree, err := newCommit.Tree()
	if err != nil {
		t.Fatalf("Failed to get tree: %v", err)
	}

	// Should have cmd/b/file.txt and cmd/b/file2.txt
	subtree, err := tree.Tree("cmd/b")
	if err != nil {
		t.Fatalf("Failed to get cmd/b subtree: %v", err)
	}

	entries := subtree.Entries
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries in cmd/b, got %d", len(entries))
	}

	// Verify commit history
	if len(newCommit.ParentHashes) != 1 {
		t.Errorf("Expected 1 parent, got %d", len(newCommit.ParentHashes))
	}

	t.Logf("Rewrite successful: %d commits rewritten, new tip: %s", res.RewrittenCommits, res.NewTip)
}

// TestPruneEmpty tests empty commit pruning
func TestPruneEmpty(t *testing.T) {
	// Create source repository
	srcDir := t.TempDir()
	srcRepo, err := git.PlainInit(srcDir, false)
	if err != nil {
		t.Fatalf("Failed to init source repo: %v", err)
	}

	wt, err := srcRepo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	sig := &object.Signature{
		Name:  "Test User",
		Email: "test@example.com",
		When:  time.Now(),
	}

	// Commit 1: Create other/file.txt (not in cmd/a)
	if err := os.MkdirAll(filepath.Join(srcDir, "other"), 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "other", "file.txt"), []byte("other"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if _, err := wt.Add("other/file.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	_, err = wt.Commit("Commit without cmd/a", &git.CommitOptions{
		Author: sig,
	})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Commit 2: Create cmd/a/file.txt
	if err := os.MkdirAll(filepath.Join(srcDir, "cmd", "a"), 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "cmd", "a", "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if _, err := wt.Add("cmd/a/file.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	_, err = wt.Commit("Add cmd/a", &git.CommitOptions{
		Author: sig,
	})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Commit 3: Modify other/file.txt (not cmd/a)
	if err := os.WriteFile(filepath.Join(srcDir, "other", "file.txt"), []byte("modified"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if _, err := wt.Add("other/file.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	commit3, err := wt.Commit("Modify other (empty for cmd/a)", &git.CommitOptions{
		Author: sig,
	})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create destination repository
	dstDir := t.TempDir()
	dstRepo, err := git.PlainInit(dstDir, true)
	if err != nil {
		t.Fatalf("Failed to init dest repo: %v", err)
	}

	// Rewrite with pruning
	rw, err := NewRewriter(srcRepo.Storer, dstRepo.Storer, Options{
		Start:      commit3,
		KeepPrefix: "cmd/a",
		NewPrefix:  "",
		PruneEmpty: true,
	})
	if err != nil {
		t.Fatalf("Failed to create rewriter: %v", err)
	}

	res, err := rw.Rewrite(context.Background())
	if err != nil {
		t.Fatalf("Rewrite failed: %v", err)
	}

	// Verify: commit1 should be pruned (empty), commit2 kept, commit3 pruned (no changes)
	if res.PrunedCommits != 2 {
		t.Errorf("Expected 2 pruned commits, got %d", res.PrunedCommits)
	}
	if res.RewrittenCommits != 1 {
		t.Errorf("Expected 1 rewritten commit, got %d", res.RewrittenCommits)
	}

	t.Logf("Prune test successful: %d commits rewritten, %d pruned", res.RewrittenCommits, res.PrunedCommits)
}

// TestRootExtraction tests extracting a subdirectory as root
func TestRootExtraction(t *testing.T) {
	// Create source repository
	srcDir := t.TempDir()
	srcRepo, err := git.PlainInit(srcDir, false)
	if err != nil {
		t.Fatalf("Failed to init source repo: %v", err)
	}

	wt, err := srcRepo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	sig := &object.Signature{
		Name:  "Test User",
		Email: "test@example.com",
		When:  time.Now(),
	}

	// Create cmd/a/main.go and cmd/a/README.md
	if err := os.MkdirAll(filepath.Join(srcDir, "cmd", "a"), 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "cmd", "a", "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "cmd", "a", "README.md"), []byte("# Project A"), 0644); err != nil {
		t.Fatalf("Failed to write README.md: %v", err)
	}
	if _, err := wt.Add("cmd/a/main.go"); err != nil {
		t.Fatalf("Failed to add main.go: %v", err)
	}
	if _, err := wt.Add("cmd/a/README.md"); err != nil {
		t.Fatalf("Failed to add README.md: %v", err)
	}

	commit1, err := wt.Commit("Initial commit", &git.CommitOptions{
		Author: sig,
	})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create destination repository
	dstDir := t.TempDir()
	dstRepo, err := git.PlainInit(dstDir, true)
	if err != nil {
		t.Fatalf("Failed to init dest repo: %v", err)
	}

	// Rewrite: extract cmd/a as root
	rw, err := NewRewriter(srcRepo.Storer, dstRepo.Storer, Options{
		Start:      commit1,
		KeepPrefix: "cmd/a",
		NewPrefix:  "", // Empty = make it root
		PruneEmpty: false,
	})
	if err != nil {
		t.Fatalf("Failed to create rewriter: %v", err)
	}

	res, err := rw.Rewrite(context.Background())
	if err != nil {
		t.Fatalf("Rewrite failed: %v", err)
	}

	// Verify tree structure
	newCommit, err := object.GetCommit(dstRepo.Storer, res.NewTip)
	if err != nil {
		t.Fatalf("Failed to get new tip commit: %v", err)
	}

	tree, err := newCommit.Tree()
	if err != nil {
		t.Fatalf("Failed to get tree: %v", err)
	}

	// Should have main.go and README.md at root
	entries := tree.Entries
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries at root, got %d", len(entries))
	}

	var foundMain, foundReadme bool
	for _, e := range entries {
		if e.Name == "main.go" {
			foundMain = true
		}
		if e.Name == "README.md" {
			foundReadme = true
		}
	}

	if !foundMain {
		t.Error("main.go not found at root")
	}
	if !foundReadme {
		t.Error("README.md not found at root")
	}

	t.Logf("Root extraction successful: %d commits rewritten", res.RewrittenCommits)
}

// TestMultipleCommits tests rewriting a longer history
func TestMultipleCommits(t *testing.T) {
	// Create source repository
	srcDir := t.TempDir()
	srcRepo, err := git.PlainInit(srcDir, false)
	if err != nil {
		t.Fatalf("Failed to init source repo: %v", err)
	}

	wt, err := srcRepo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	sig := &object.Signature{
		Name:  "Test User",
		Email: "test@example.com",
		When:  time.Now(),
	}

	// Create directory structure
	if err := os.MkdirAll(filepath.Join(srcDir, "cmd", "a"), 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	var lastCommit plumbing.Hash

	// Create 5 commits
	for i := 1; i <= 5; i++ {
		filename := filepath.Join(srcDir, "cmd", "a", fmt.Sprintf("file%d.txt", i))
		content := fmt.Sprintf("content %d", i)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file%d: %v", i, err)
		}
		if _, err := wt.Add(fmt.Sprintf("cmd/a/file%d.txt", i)); err != nil {
			t.Fatalf("Failed to add file%d: %v", i, err)
		}
		lastCommit, err = wt.Commit(fmt.Sprintf("Commit %d", i), &git.CommitOptions{
			Author: sig,
		})
		if err != nil {
			t.Fatalf("Failed to commit %d: %v", i, err)
		}
	}

	// Create destination repository
	dstDir := t.TempDir()
	dstRepo, err := git.PlainInit(dstDir, true)
	if err != nil {
		t.Fatalf("Failed to init dest repo: %v", err)
	}

	// Rewrite all 5 commits
	rw, err := NewRewriter(srcRepo.Storer, dstRepo.Storer, Options{
		Start:      lastCommit,
		KeepPrefix: "cmd/a",
		NewPrefix:  "pkg/b",
		PruneEmpty: false,
	})
	if err != nil {
		t.Fatalf("Failed to create rewriter: %v", err)
	}

	res, err := rw.Rewrite(context.Background())
	if err != nil {
		t.Fatalf("Rewrite failed: %v", err)
	}

	if res.RewrittenCommits != 5 {
		t.Errorf("Expected 5 rewritten commits, got %d", res.RewrittenCommits)
	}

	// Verify final tree has all 5 files
	newCommit, err := object.GetCommit(dstRepo.Storer, res.NewTip)
	if err != nil {
		t.Fatalf("Failed to get new tip commit: %v", err)
	}

	tree, err := newCommit.Tree()
	if err != nil {
		t.Fatalf("Failed to get tree: %v", err)
	}

	subtree, err := tree.Tree("pkg/b")
	if err != nil {
		t.Fatalf("Failed to get pkg/b subtree: %v", err)
	}

	if len(subtree.Entries) != 5 {
		t.Errorf("Expected 5 files in pkg/b, got %d", len(subtree.Entries))
	}

	t.Logf("Multiple commits test successful: %d commits rewritten", res.RewrittenCommits)
}

// TestWalkerOrdering tests that the walker produces parent-first ordering
func TestWalkerOrdering(t *testing.T) {
	// Create source repository
	srcDir := t.TempDir()
	srcRepo, err := git.PlainInit(srcDir, false)
	if err != nil {
		t.Fatalf("Failed to init source repo: %v", err)
	}

	wt, err := srcRepo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	sig := &object.Signature{
		Name:  "Test User",
		Email: "test@example.com",
		When:  time.Now(),
	}

	// Create a simple linear history
	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("1"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if _, err := wt.Add("file.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	commit1, err := wt.Commit("Commit 1", &git.CommitOptions{Author: sig})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("2"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if _, err := wt.Add("file.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	commit2, err := wt.Commit("Commit 2", &git.CommitOptions{Author: sig})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("3"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if _, err := wt.Add("file.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	commit3, err := wt.Commit("Commit 3", &git.CommitOptions{Author: sig})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Walk from commit3
	walker := CommitTopoWalker{Src: srcRepo.Storer}
	order, err := walker.WalkFrom(context.Background(), commit3)
	if err != nil {
		t.Fatalf("Walker failed: %v", err)
	}

	// Should be [commit1, commit2, commit3]
	if len(order) != 3 {
		t.Errorf("Expected 3 commits, got %d", len(order))
	}

	if order[0] != commit1 {
		t.Errorf("Expected first commit to be commit1, got %s", order[0])
	}
	if order[1] != commit2 {
		t.Errorf("Expected second commit to be commit2, got %s", order[1])
	}
	if order[2] != commit3 {
		t.Errorf("Expected third commit to be commit3, got %s", order[2])
	}

	t.Log("Walker ordering test successful")
}
