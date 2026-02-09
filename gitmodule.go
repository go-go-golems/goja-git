package main

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitModule provides the top-level git object for JS
type GitModule struct {
	rt *goja.Runtime
}

// OpenOptions for git.open() - use exported fields matching JS property names
type OpenOptions struct {
	Dir string // matches { Dir: "..." } in JS
}

// InitOptions for git.init()
type InitOptions struct {
	Dir           string
	DefaultBranch string
	Bare          bool
}

// AddOptions for repo.add()
type AddOptions struct {
	Paths []string
	All   bool
}

// CommitOptions for repo.commit()
type CommitOptions struct {
	Message string
	Author  struct {
		Name  string
		Email string
	}
	Amend bool
}

// LogOptions for repo.log()
type LogOptions struct {
	Ref   string
	Depth int
}

// CheckoutOptions for repo.checkout()
type CheckoutOptions struct {
	Ref    string
	Create bool
	Force  bool
}

// BranchCreateOptions for repo.branch.create()
type BranchCreateOptions struct {
	Name       string
	StartPoint string
}

// DiffOptions for repo.diff()
type DiffOptions struct {
	From  string
	To    string
	Paths []string
}

// TagOptions for repo.tag.create()
type TagOptions struct {
	Name    string
	Message string
	Ref     string
}

// RepoHandle represents a git repository exposed to JS
type RepoHandle struct {
	rt   *goja.Runtime
	dir  string
	repo *git.Repository
}

// InstallGit installs the git module into the goja runtime
func InstallGit(rt *goja.Runtime) {
	m := &GitModule{rt: rt}

	gitObj := rt.NewObject()
	_ = gitObj.Set("open", m.Open)
	_ = gitObj.Set("init", m.Init)

	rt.Set("git", gitObj)
}

// Open opens an existing git repository
func (m *GitModule) Open(call goja.FunctionCall) goja.Value {
	var opts OpenOptions
	m.mustExport(call.Argument(0), &opts)

	if opts.Dir == "" {
		panic(m.rt.NewGoError(fmt.Errorf("Dir is required")))
	}

	repo, err := git.PlainOpen(opts.Dir)
	if err != nil {
		panic(m.rt.NewGoError(err))
	}
	return m.newRepoValue(opts.Dir, repo)
}

// Init initializes a new git repository
func (m *GitModule) Init(call goja.FunctionCall) goja.Value {
	var opts InitOptions
	m.mustExport(call.Argument(0), &opts)

	if opts.Dir == "" {
		panic(m.rt.NewGoError(fmt.Errorf("Dir is required")))
	}

	repo, err := git.PlainInit(opts.Dir, opts.Bare)
	if err != nil {
		panic(m.rt.NewGoError(err))
	}

	// Set default branch if specified
	if opts.DefaultBranch != "" && !opts.Bare {
		// Create an initial commit to establish the branch
		wt, err := repo.Worktree()
		if err == nil {
			// Try to create the branch reference
			headRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.ReferenceName("refs/heads/"+opts.DefaultBranch))
			err = repo.Storer.SetReference(headRef)
			if err != nil {
				// Non-fatal, just log
				fmt.Printf("Warning: could not set default branch: %v\n", err)
			}
			_ = wt
		}
	}

	return m.newRepoValue(opts.Dir, repo)
}

// newRepoValue creates a JS object representing a git repository
func (m *GitModule) newRepoValue(dir string, repo *git.Repository) goja.Value {
	h := &RepoHandle{rt: m.rt, dir: dir, repo: repo}

	o := m.rt.NewObject()
	_ = o.Set("dir", dir)

	// Porcelain methods
	_ = o.Set("status", h.Status)
	_ = o.Set("add", h.Add)
	_ = o.Set("commit", h.Commit)
	_ = o.Set("log", h.Log)
	_ = o.Set("checkout", h.Checkout)
	_ = o.Set("diff", h.Diff)

	// Branch operations
	branchObj := m.rt.NewObject()
	_ = branchObj.Set("list", h.BranchList)
	_ = branchObj.Set("create", h.BranchCreate)
	_ = branchObj.Set("current", h.BranchCurrent)
	_ = o.Set("branch", branchObj)

	// Tag operations
	tagObj := m.rt.NewObject()
	_ = tagObj.Set("list", h.TagList)
	_ = tagObj.Set("create", h.TagCreate)
	_ = o.Set("tag", tagObj)

	// Plumbing methods
	refsObj := m.rt.NewObject()
	_ = refsObj.Set("resolve", h.RefsResolve)
	_ = o.Set("refs", refsObj)

	return o
}

// Status returns the working tree status
func (h *RepoHandle) Status(call goja.FunctionCall) goja.Value {
	wt, err := h.repo.Worktree()
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	st, err := wt.Status()
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	// Convert go-git status map to JS array of objects
	out := make([]map[string]any, 0, len(st))
	for path, fs := range st {
		out = append(out, map[string]any{
			"path":     path,
			"staging":  string(fs.Staging),
			"worktree": string(fs.Worktree),
		})
	}
	return h.rt.ToValue(out)
}

// Add adds files to the staging area
func (h *RepoHandle) Add(call goja.FunctionCall) goja.Value {
	var opts AddOptions
	h.mustExport(call.Argument(0), &opts)

	wt, err := h.repo.Worktree()
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	if opts.All {
		if _, err := wt.Add("."); err != nil {
			panic(h.rt.NewGoError(err))
		}
	} else {
		for _, p := range opts.Paths {
			if _, err := wt.Add(p); err != nil {
				panic(h.rt.NewGoError(err))
			}
		}
	}
	return goja.Undefined()
}

// Commit creates a new commit
func (h *RepoHandle) Commit(call goja.FunctionCall) goja.Value {
	var opts CommitOptions
	h.mustExport(call.Argument(0), &opts)

	if opts.Message == "" {
		panic(h.rt.NewGoError(fmt.Errorf("Message is required")))
	}

	wt, err := h.repo.Worktree()
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	authorName := opts.Author.Name
	authorEmail := opts.Author.Email
	if authorName == "" {
		authorName = "Unknown"
	}
	if authorEmail == "" {
		authorEmail = "unknown@example.com"
	}

	co := &git.CommitOptions{
		All: false,
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
		Amend: opts.Amend,
	}

	hash, err := wt.Commit(opts.Message, co)
	if err != nil {
		panic(h.rt.NewGoError(err))
	}
	return h.rt.ToValue(hash.String())
}

// Log retrieves commit history
func (h *RepoHandle) Log(call goja.FunctionCall) goja.Value {
	var opts LogOptions
	if !goja.IsUndefined(call.Argument(0)) && !goja.IsNull(call.Argument(0)) {
		h.mustExport(call.Argument(0), &opts)
	}
	if opts.Depth <= 0 {
		opts.Depth = 50
	}
	if opts.Ref == "" {
		opts.Ref = "HEAD"
	}

	revHash, err := h.resolveRef(opts.Ref)
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	iter, err := h.repo.Log(&git.LogOptions{From: revHash})
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	out := make([]map[string]any, 0, opts.Depth)
	i := 0
	_ = iter.ForEach(func(c *object.Commit) error {
		if i >= opts.Depth {
			return fmt.Errorf("stop") // stop iteration
		}
		out = append(out, map[string]any{
			"oid":     c.Hash.String(),
			"message": c.Message,
			"author": map[string]any{
				"name":  c.Author.Name,
				"email": c.Author.Email,
				"when":  c.Author.When.Format(time.RFC3339),
			},
		})
		i++
		return nil
	})

	return h.rt.ToValue(out)
}

// Checkout switches branches or restores working tree files
func (h *RepoHandle) Checkout(call goja.FunctionCall) goja.Value {
	var opts CheckoutOptions
	h.mustExport(call.Argument(0), &opts)

	if opts.Ref == "" {
		panic(h.rt.NewGoError(fmt.Errorf("Ref is required")))
	}

	wt, err := h.repo.Worktree()
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	checkoutOpts := &git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + opts.Ref),
		Create: opts.Create,
		Force:  opts.Force,
	}

	if err := wt.Checkout(checkoutOpts); err != nil {
		panic(h.rt.NewGoError(err))
	}

	return goja.Undefined()
}

// Diff shows changes between commits, commit and working tree, etc
func (h *RepoHandle) Diff(call goja.FunctionCall) goja.Value {
	var opts DiffOptions
	h.mustExport(call.Argument(0), &opts)

	// Simple implementation: return file names that changed
	// For a full diff, we'd need to implement patch generation

	if opts.From == "" {
		opts.From = "HEAD"
	}
	if opts.To == "" {
		opts.To = "HEAD"
	}

	fromHash, err := h.resolveRef(opts.From)
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	toHash, err := h.resolveRef(opts.To)
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	fromCommit, err := h.repo.CommitObject(fromHash)
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	toCommit, err := h.repo.CommitObject(toHash)
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	fromTree, err := fromCommit.Tree()
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	toTree, err := toCommit.Tree()
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	changes, err := fromTree.Diff(toTree)
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	out := make([]map[string]any, 0, len(changes))
	for _, change := range changes {
		from, to, err := change.Files()
		if err != nil {
			continue
		}

		fromName := ""
		toName := ""
		if from != nil {
			fromName = from.Name
		}
		if to != nil {
			toName = to.Name
		}

		out = append(out, map[string]any{
			"from": fromName,
			"to":   toName,
		})
	}

	return h.rt.ToValue(out)
}

// BranchList lists all branches
func (h *RepoHandle) BranchList(call goja.FunctionCall) goja.Value {
	refs, err := h.repo.Branches()
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	out := make([]string, 0)
	_ = refs.ForEach(func(ref *plumbing.Reference) error {
		out = append(out, ref.Name().Short())
		return nil
	})

	return h.rt.ToValue(out)
}

// BranchCreate creates a new branch
func (h *RepoHandle) BranchCreate(call goja.FunctionCall) goja.Value {
	var opts BranchCreateOptions
	h.mustExport(call.Argument(0), &opts)

	if opts.Name == "" {
		panic(h.rt.NewGoError(fmt.Errorf("Name is required")))
	}

	startPoint := opts.StartPoint
	if startPoint == "" {
		startPoint = "HEAD"
	}

	hash, err := h.resolveRef(startPoint)
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	refName := plumbing.NewBranchReferenceName(opts.Name)
	ref := plumbing.NewHashReference(refName, hash)

	if err := h.repo.Storer.SetReference(ref); err != nil {
		panic(h.rt.NewGoError(err))
	}

	return goja.Undefined()
}

// BranchCurrent returns the current branch name
func (h *RepoHandle) BranchCurrent(call goja.FunctionCall) goja.Value {
	head, err := h.repo.Head()
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	if !head.Name().IsBranch() {
		return h.rt.ToValue("(detached)")
	}

	return h.rt.ToValue(head.Name().Short())
}

// TagList lists all tags
func (h *RepoHandle) TagList(call goja.FunctionCall) goja.Value {
	refs, err := h.repo.Tags()
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	out := make([]string, 0)
	_ = refs.ForEach(func(ref *plumbing.Reference) error {
		out = append(out, ref.Name().Short())
		return nil
	})

	return h.rt.ToValue(out)
}

// TagCreate creates a new tag
func (h *RepoHandle) TagCreate(call goja.FunctionCall) goja.Value {
	var opts TagOptions
	h.mustExport(call.Argument(0), &opts)

	if opts.Name == "" {
		panic(h.rt.NewGoError(fmt.Errorf("Name is required")))
	}

	targetRef := opts.Ref
	if targetRef == "" {
		targetRef = "HEAD"
	}

	hash, err := h.resolveRef(targetRef)
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	refName := plumbing.NewTagReferenceName(opts.Name)
	ref := plumbing.NewHashReference(refName, hash)

	if err := h.repo.Storer.SetReference(ref); err != nil {
		panic(h.rt.NewGoError(err))
	}

	return h.rt.ToValue(hash.String())
}

// RefsResolve resolves a reference to a hash
func (h *RepoHandle) RefsResolve(call goja.FunctionCall) goja.Value {
	var opts struct {
		Ref string
	}
	h.mustExport(call.Argument(0), &opts)

	if opts.Ref == "" {
		panic(h.rt.NewGoError(fmt.Errorf("Ref is required")))
	}

	hash, err := h.resolveRef(opts.Ref)
	if err != nil {
		panic(h.rt.NewGoError(err))
	}

	return h.rt.ToValue(hash.String())
}

// resolveRef resolves a reference name to a hash
func (h *RepoHandle) resolveRef(ref string) (plumbing.Hash, error) {
	// Try as hash first
	if len(ref) >= 7 && len(ref) <= 40 {
		hash := plumbing.NewHash(ref)
		if hash != plumbing.ZeroHash {
			// Verify it exists
			_, err := h.repo.CommitObject(hash)
			if err == nil {
				return hash, nil
			}
		}
	}

	// Try as revision
	rev := plumbing.Revision(ref)
	hash, err := h.repo.ResolveRevision(rev)
	if err != nil {
		return plumbing.ZeroHash, err
	}
	return *hash, nil
}

// mustExport exports a goja value to a Go struct, panicking on error
func (m *GitModule) mustExport(v goja.Value, dst any) {
	if goja.IsUndefined(v) || goja.IsNull(v) {
		// Don't export undefined/null values
		return
	}
	if err := m.rt.ExportTo(v, dst); err != nil {
		panic(m.rt.NewGoError(err))
	}
}

// mustExport exports a goja value to a Go struct, panicking on error
func (h *RepoHandle) mustExport(v goja.Value, dst any) {
	if goja.IsUndefined(v) || goja.IsNull(v) {
		// Don't export undefined/null values
		return
	}
	if err := h.rt.ExportTo(v, dst); err != nil {
		panic(h.rt.NewGoError(err))
	}
}
