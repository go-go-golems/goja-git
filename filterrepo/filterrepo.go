package filterrepo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// Options covers the subset of git-filter-repo semantics:
//   - keep a directory prefix (like "cmd/a")
//   - rewrite it to a new prefix (like "cmd/b") in EVERY rewritten commit
//   - optionally prune empty commits
type Options struct {
	// Start is the tip commit hash you want to rewrite from (reachable history will be rewritten).
	Start plumbing.Hash

	// KeepPrefix is the directory path to keep from the source commit trees (e.g. "cmd/a").
	KeepPrefix string

	// NewPrefix is where to place the kept subtree in the rewritten commits (e.g. "cmd/b").
	// If empty, the kept subtree becomes the rewritten root tree.
	NewPrefix string

	// PruneEmpty prunes commits whose rewritten tree equals their (rewritten) first parent tree.
	// This is conservative by default: merge commits are NOT pruned unless PruneMerges is true.
	PruneEmpty bool

	// PruneMerges allows pruning merge commits when the rewritten tree matches first parent.
	// This can change topology; keep false unless you explicitly want it.
	PruneMerges bool
}

type Result struct {
	NewTip           plumbing.Hash
	RewrittenCommits int
	PrunedCommits    int
}

// --- Subcomponent: commit topo walker (parents before children) ---

type CommitTopoWalker struct {
	Src storer.EncodedObjectStorer
}

func (w CommitTopoWalker) WalkFrom(ctx context.Context, start plumbing.Hash) ([]plumbing.Hash, error) {
	visited := map[plumbing.Hash]bool{}
	order := make([]plumbing.Hash, 0, 4096)

	type frame struct {
		h        plumbing.Hash
		expanded bool
	}

	stack := []frame{{h: start, expanded: false}}
	for len(stack) > 0 {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		}

		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if f.expanded {
			order = append(order, f.h)
			continue
		}
		if visited[f.h] {
			continue
		}
		visited[f.h] = true

		// Post-order: push self-as-expanded, then parents.
		stack = append(stack, frame{h: f.h, expanded: true})

		c, err := object.GetCommit(w.Src, f.h)
		if err != nil {
			return nil, err
		}
		for _, p := range c.ParentHashes {
			if !visited[p] {
				stack = append(stack, frame{h: p, expanded: false})
			}
		}
	}

	// order is parent-first (topologically valid for rewriting with parent remap).
	return order, nil
}

// --- Subcomponent: object copier (trees/blobs) ---

type ObjectCopier struct {
	Src storer.EncodedObjectStorer
	Dst storer.EncodedObjectStorer

	copiedTrees map[plumbing.Hash]struct{}
	copiedBlobs map[plumbing.Hash]struct{}
}

func NewObjectCopier(src, dst storer.EncodedObjectStorer) *ObjectCopier {
	return &ObjectCopier{
		Src:         src,
		Dst:         dst,
		copiedTrees: map[plumbing.Hash]struct{}{},
		copiedBlobs: map[plumbing.Hash]struct{}{},
	}
}

func (c *ObjectCopier) CopyBlob(hash plumbing.Hash) error {
	if hash == (plumbing.Hash{}) {
		return nil
	}
	if _, ok := c.copiedBlobs[hash]; ok {
		return nil
	}

	// If already in destination, consider copied.
	if _, err := c.Dst.EncodedObject(plumbing.BlobObject, hash); err == nil {
		c.copiedBlobs[hash] = struct{}{}
		return nil
	} else if err != nil && !errors.Is(err, plumbing.ErrObjectNotFound) {
		return err
	}

	srcObj, err := c.Src.EncodedObject(plumbing.BlobObject, hash)
	if err != nil {
		return err
	}

	dstObj := &plumbing.MemoryObject{}
	dstObj.SetType(plumbing.BlobObject)
	dstObj.SetSize(srcObj.Size())

	r, err := srcObj.Reader()
	if err != nil {
		return err
	}
	defer r.Close()

	wr, err := dstObj.Writer()
	if err != nil {
		return err
	}
	if _, err := io.Copy(wr, r); err != nil {
		_ = wr.Close()
		return err
	}
	if err := wr.Close(); err != nil {
		return err
	}

	newHash, err := c.Dst.SetEncodedObject(dstObj)
	if err != nil {
		return err
	}
	if newHash != hash {
		return fmt.Errorf("blob hash changed while copying: src=%s dst=%s", hash, newHash)
	}

	c.copiedBlobs[hash] = struct{}{}
	return nil
}

func (c *ObjectCopier) CopyTree(hash plumbing.Hash) error {
	if hash == (plumbing.Hash{}) {
		return nil
	}
	if _, ok := c.copiedTrees[hash]; ok {
		return nil
	}

	// If already in destination, consider copied.
	if _, err := c.Dst.EncodedObject(plumbing.TreeObject, hash); err == nil {
		c.copiedTrees[hash] = struct{}{}
		return nil
	} else if err != nil && !errors.Is(err, plumbing.ErrObjectNotFound) {
		return err
	}

	srcObj, err := c.Src.EncodedObject(plumbing.TreeObject, hash)
	if err != nil {
		return err
	}

	var t object.Tree
	if err := t.Decode(srcObj); err != nil {
		return err
	}

	// Ensure children exist in destination.
	for _, e := range t.Entries {
		switch e.Mode {
		case filemode.Dir:
			if err := c.CopyTree(e.Hash); err != nil {
				return err
			}
		case filemode.Submodule:
			// gitlink entry: hash points at a commit in another repo; do not copy objects.
		default:
			if err := c.CopyBlob(e.Hash); err != nil {
				return err
			}
		}
	}

	// Store the tree object itself in destination.
	dstObj := &plumbing.MemoryObject{}
	dstObj.SetType(plumbing.TreeObject)
	if err := t.Encode(dstObj); err != nil {
		return err
	}

	newHash, err := c.Dst.SetEncodedObject(dstObj)
	if err != nil {
		return err
	}
	if newHash != hash {
		return fmt.Errorf("tree hash changed while copying: src=%s dst=%s", hash, newHash)
	}

	c.copiedTrees[hash] = struct{}{}
	return nil
}

// --- Subcomponent: tree builder (mount a leaf tree at a new prefix) ---

type TreeBuilder struct {
	Dst storer.EncodedObjectStorer
}

func (b TreeBuilder) WriteTree(entries []object.TreeEntry) (plumbing.Hash, error) {
	t := &object.Tree{Entries: entries}

	obj := &plumbing.MemoryObject{}
	obj.SetType(plumbing.TreeObject)
	if err := t.Encode(obj); err != nil {
		return plumbing.Hash{}, err
	}

	return b.Dst.SetEncodedObject(obj)
}

// BuildPrefixedRoot creates a brand-new root tree that contains ONLY `prefix`
// leading to `leafTreeHash`.
//
// Example: prefix="cmd/b" and leafTreeHash=<hash of subtree>
//   root:  {"cmd" -> tree1}
//   tree1: {"b"   -> leafTreeHash}
func (b TreeBuilder) BuildPrefixedRoot(prefix string, leafTreeHash plumbing.Hash) (plumbing.Hash, error) {
	prefix = normPath(prefix)
	if prefix == "" {
		return leafTreeHash, nil
	}

	parts := strings.Split(prefix, "/")
	cur := leafTreeHash

	// Build upward: (.../cmd/b) means root has "cmd" which has "b" which points at leaf.
	for i := len(parts) - 1; i >= 0; i-- {
		h, err := b.WriteTree([]object.TreeEntry{
			{
				Name: parts[i],
				Mode: filemode.Dir,
				Hash: cur,
			},
		})
		if err != nil {
			return plumbing.Hash{}, err
		}
		cur = h
	}

	return cur, nil
}

// --- The main rewriter (ties walker + copier + tree builder together) ---

type Rewriter struct {
	Src  storer.EncodedObjectStorer
	Dst  storer.EncodedObjectStorer
	Opts Options

	walker      CommitTopoWalker
	copier      *ObjectCopier
	treeb       TreeBuilder
	emptyTree   plumbing.Hash
	emptyInited bool

	byOld map[plumbing.Hash]rewriteInfo
}

type rewriteInfo struct {
	NewCommit plumbing.Hash
	NewTree   plumbing.Hash
}

func NewRewriter(src, dst storer.EncodedObjectStorer, opts Options) (*Rewriter, error) {
	if opts.Start == (plumbing.Hash{}) {
		return nil, fmt.Errorf("Options.Start is required")
	}
	opts.KeepPrefix = normPath(opts.KeepPrefix)
	opts.NewPrefix = normPath(opts.NewPrefix)

	r := &Rewriter{
		Src:    src,
		Dst:    dst,
		Opts:   opts,
		walker: CommitTopoWalker{Src: src},
		copier: NewObjectCopier(src, dst),
		treeb:  TreeBuilder{Dst: dst},
		byOld:  map[plumbing.Hash]rewriteInfo{},
	}
	return r, nil
}

func (r *Rewriter) Rewrite(ctx context.Context) (Result, error) {
	order, err := r.walker.WalkFrom(ctx, r.Opts.Start)
	if err != nil {
		return Result{}, err
	}

	var res Result
	emptyTree, err := r.emptyTreeHash()
	if err != nil {
		return Result{}, err
	}

	for _, oldHash := range order {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return Result{}, ctx.Err()
			default:
			}
		}

		oldCommit, err := object.GetCommit(r.Src, oldHash)
		if err != nil {
			return Result{}, err
		}

		newParents, firstParentInfo := r.remapParents(oldCommit.ParentHashes)

		leafTreeHash, leafOK, err := r.leafTree(oldCommit)
		if err != nil {
			return Result{}, err
		}

		var newTree plumbing.Hash
		if !leafOK {
			// Path absent => output tree is empty (no cmd/b directory at all).
			newTree = emptyTree
		} else {
			// Ensure leaf subtree exists in destination (and its blobs/trees).
			if err := r.copier.CopyTree(leafTreeHash); err != nil {
				return Result{}, err
			}

			if r.Opts.NewPrefix == "" {
				newTree = leafTreeHash
			} else {
				h, err := r.treeb.BuildPrefixedRoot(r.Opts.NewPrefix, leafTreeHash)
				if err != nil {
					return Result{}, err
				}
				newTree = h
			}
		}

		// Conservative prune-empty.
		if r.Opts.PruneEmpty {
			if len(newParents) == 0 {
				// If we're a would-be root and empty, just drop it (map to "no commit").
				if newTree == emptyTree {
					r.byOld[oldHash] = rewriteInfo{NewCommit: plumbing.Hash{}, NewTree: emptyTree}
					res.PrunedCommits++
					continue
				}
			} else {
				allow := (len(newParents) == 1) || r.Opts.PruneMerges
				if allow && firstParentInfo.NewTree != (plumbing.Hash{}) && newTree == firstParentInfo.NewTree {
					// Drop this commit by mapping it to its first parent rewrite.
					r.byOld[oldHash] = firstParentInfo
					res.PrunedCommits++
					continue
				}
			}
		}

		// Write new commit object.
		newCommit := &object.Commit{
			Author:       oldCommit.Author,
			Committer:    oldCommit.Committer,
			Message:      oldCommit.Message,
			TreeHash:     newTree,
			ParentHashes: newParents,

			// Note: signatures don't remain valid across rewrites.
			PGPSignature: "",
		}

		obj := &plumbing.MemoryObject{}
		obj.SetType(plumbing.CommitObject)
		if err := newCommit.Encode(obj); err != nil {
			return Result{}, err
		}

		newHash, err := r.Dst.SetEncodedObject(obj)
		if err != nil {
			return Result{}, err
		}

		r.byOld[oldHash] = rewriteInfo{NewCommit: newHash, NewTree: newTree}
		res.RewrittenCommits++
	}

	info, ok := r.byOld[r.Opts.Start]
	if !ok || info.NewCommit == (plumbing.Hash{}) {
		return Result{}, fmt.Errorf("rewrite produced no tip commit (path may never have existed?)")
	}

	res.NewTip = info.NewCommit
	return res, nil
}

func (r *Rewriter) emptyTreeHash() (plumbing.Hash, error) {
	if r.emptyInited {
		return r.emptyTree, nil
	}
	h, err := r.treeb.WriteTree(nil)
	if err != nil {
		return plumbing.Hash{}, err
	}
	r.emptyTree = h
	r.emptyInited = true
	return h, nil
}

func (r *Rewriter) remapParents(oldParents []plumbing.Hash) (newParents []plumbing.Hash, firstParent rewriteInfo) {
	seen := map[plumbing.Hash]bool{}
	out := make([]plumbing.Hash, 0, len(oldParents))

	var firstSet bool
	for _, p := range oldParents {
		pi, ok := r.byOld[p]
		if !ok {
			// With correct topo-walk, parents should already be processed.
			continue
		}
		if pi.NewCommit == (plumbing.Hash{}) {
			// Parent got pruned to "no commit" (root empty chain), drop it.
			continue
		}
		if !firstSet {
			firstParent = pi
			firstSet = true
		}
		if !seen[pi.NewCommit] {
			seen[pi.NewCommit] = true
			out = append(out, pi.NewCommit)
		}
	}

	return out, firstParent
}

func (r *Rewriter) leafTree(c *object.Commit) (plumbing.Hash, bool, error) {
	root, err := c.Tree()
	if err != nil {
		return plumbing.Hash{}, false, err
	}

	if r.Opts.KeepPrefix == "" {
		return root.Hash, true, nil
	}

	sub, err := root.Tree(r.Opts.KeepPrefix)
	if err != nil {
		// Expected for commits where the directory doesn't exist yet.
		if errors.Is(err, object.ErrDirectoryNotFound) || errors.Is(err, object.ErrFileNotFound) {
			return plumbing.Hash{}, false, nil
		}
		return plumbing.Hash{}, false, err
	}
	return sub.Hash, true, nil
}

func normPath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.TrimSpace(p)
	p = strings.Trim(p, "/")
	if p == "." {
		return ""
	}
	return p
}
