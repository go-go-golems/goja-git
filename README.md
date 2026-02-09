# Goja Git Wrapper

A JavaScript wrapper for Git operations using [goja](https://github.com/dop251/goja) (Go JavaScript VM) and [go-git](https://github.com/go-git/go-git).

## Overview

This project provides a JavaScript API for Git operations that can be executed in a Go-embedded JavaScript runtime. It allows you to write Git automation scripts in JavaScript while leveraging the performance and safety of Go.

## Features

- **Repository Management**: Initialize and open Git repositories
- **Staging & Commits**: Add files and create commits
- **Branching**: Create, list, and checkout branches
- **History**: View commit logs and resolve references
- **Tags**: Create and list tags
- **Diff**: Compare changes between commits
- **Filter-Repo**: Rewrite repository history to extract subdirectories

## Installation

### Prerequisites

- Go 1.25.5 or later
- Git (for testing)

### Build

```bash
go build -o goja-git .
```

## Usage

### Running Scripts

```bash
./goja-git <script.js>
```

### JavaScript API

The API is designed to be similar to popular JavaScript Git libraries like `simple-git` and `isomorphic-git`, with a focus on simplicity and clarity.

#### Initialize or Open a Repository

```javascript
// Initialize a new repository
const repo = git.init({
  Dir: "/path/to/repo",
  DefaultBranch: "main",
  Bare: false
});

// Open an existing repository
const repo = git.open({
  Dir: "/path/to/repo"
});
```

#### Check Status

```javascript
const status = repo.status();
// Returns array of: [{ path: "file.txt", staging: "M", worktree: " " }]
```

#### Add Files

```javascript
// Add specific files
repo.add({
  Paths: ["file1.txt", "file2.txt"]
});

// Add all files
repo.add({
  All: true
});
```

#### Create Commits

```javascript
const commitHash = repo.commit({
  Message: "feat: add new feature",
  Author: {
    Name: "John Doe",
    Email: "john@example.com"
  },
  Amend: false
});
```

#### View Commit History

```javascript
const commits = repo.log({
  Ref: "HEAD",
  Depth: 10
});

commits.forEach(commit => {
  console.log(commit.oid, commit.message);
  console.log("Author:", commit.author.name, commit.author.email);
});
```

#### Branch Operations

```javascript
// List branches
const branches = repo.branch.list();

// Get current branch
const current = repo.branch.current();

// Create a new branch
repo.branch.create({
  Name: "feature-branch",
  StartPoint: "HEAD"
});

// Checkout a branch
repo.checkout({
  Ref: "feature-branch",
  Create: false,
  Force: false
});
```

#### Tag Operations

```javascript
// List tags
const tags = repo.tag.list();

// Create a tag
const tagHash = repo.tag.create({
  Name: "v1.0.0",
  Message: "Release 1.0.0",
  Ref: "HEAD"
});
```

#### Diff Operations

```javascript
const changes = repo.diff({
  From: "HEAD~1",
  To: "HEAD"
});

changes.forEach(change => {
  console.log("Changed:", change.from, "->", change.to);
});
```

#### Reference Resolution

```javascript
const headHash = repo.refs.resolve({
  Ref: "HEAD"
});

const branchHash = repo.refs.resolve({
  Ref: "main"
});
```

#### Filter-Repo Operations

```javascript
// Extract a subdirectory as a new repository
const filtered = repo.filterRepo({
  OutDir: "/path/to/output",
  Ref: "HEAD",
  Path: "cmd/a",           // Keep only this path
  ToPrefix: "cmd/b",        // Rename to this prefix (empty = root)
  PruneEmpty: true,         // Remove commits that don't touch the path
  PruneMerges: false,       // Keep merge commits even if empty
  OutBranch: "main"         // Branch name in output repo
});

console.log("New tip:", filtered.newTip);
console.log("Rewritten commits:", filtered.rewrittenCommits);
console.log("Pruned commits:", filtered.prunedCommits);
```

**Use Cases:**

- Extract a subdirectory from a monorepo into its own repository
- Rename directory paths throughout history
- Clean up history by removing commits unrelated to a specific path
- Split a large repository into smaller focused repositories

## Example Scripts

The `scripts/` directory contains several example scripts demonstrating different Git operations:

1. **01-init-and-commit.js** - Initialize a repository and create commits
2. **02-branching.js** - Branch creation and checkout operations
3. **03-log-and-history.js** - View commit history and resolve references
4. **04-tags-and-diff.js** - Tag management and diff operations
5. **05-complete-workflow.js** - Complete workflow demonstration

### Filter-Repo Examples

The `scripts-filterrepo/` directory contains comprehensive filter-repo tests:

1. **01-basic-filter-rename.js** - Basic path filtering and renaming
2. **02-extract-to-root.js** - Extract subdirectory as repository root
3. **03-prune-empty.js** - Prune commits without target path
4. **04-deep-paths.js** - Filter deeply nested directory structures
5. **05-complete-workflow.js** - Extract project from monorepo

### Running Examples

```bash
# Initialize and make first commit
./goja-git scripts/01-init-and-commit.js

# Work with branches
./goja-git scripts/02-branching.js

# View history
./goja-git scripts/03-log-and-history.js

# Tags and diffs
./goja-git scripts/04-tags-and-diff.js

# Complete workflow
./goja-git scripts/05-complete-workflow.js
```

## API Design Notes

### Property Name Capitalization

The JavaScript API uses **capitalized property names** (e.g., `Dir`, `Message`, `Name`) to match Go struct field names. This is a requirement of goja's `ExportTo` function, which performs exact case-sensitive matching.

**Example:**

```javascript
// Correct
git.open({ Dir: "/path" })

// Incorrect (will not work)
git.open({ dir: "/path" })
```

### Options Objects

All methods accept a single options object parameter, following the pattern of `isomorphic-git`. This makes the API easy to extend without breaking existing code.

### Error Handling

Errors are thrown as JavaScript exceptions and can be caught with try/catch:

```javascript
try {
  const repo = git.open({ Dir: "/nonexistent" });
} catch (e) {
  console.log("Error:", e.message);
}
```

## Architecture

### Components

- **gitmodule.go** - Core Git wrapper module exposing JavaScript API
- **filterrepo/filterrepo.go** - Git filter-repo implementation
- **filterrepo/filterrepo_test.go** - Comprehensive Go tests
- **main.go** - CLI runner that executes JavaScript files
- **scripts/** - Example JavaScript scripts
- **scripts-filterrepo/** - Filter-repo test scripts

### Technology Stack

- **goja** - ECMAScript 5.1+ implementation in Go
- **go-git** - Pure Go implementation of Git
- **Go 1.25.5** - Programming language

## Testing

### Basic Git Operations

A test repository is included in `test-repo/` for running the example scripts. The scripts are designed to be run sequentially to demonstrate a complete Git workflow.

### Filter-Repo Tests

**Go Tests:**

```bash
go test -v ./filterrepo/
```

The Go test suite includes:

- Basic path filtering and renaming
- Empty commit pruning
- Root extraction
- Multiple commit history rewriting
- Topological commit ordering

**JavaScript Tests:**

```bash
# Setup test files
./scripts-filterrepo/setup-test-files.sh

# Run individual tests
./goja-git scripts-filterrepo/01-basic-filter-rename.js
./goja-git scripts-filterrepo/02-extract-to-root.js
./goja-git scripts-filterrepo/03-prune-empty.js
./goja-git scripts-filterrepo/04-deep-paths.js
./goja-git scripts-filterrepo/05-complete-workflow.js
```

## Limitations

- No remote operations (fetch, push, pull) - can be added in future versions
- Diff operations return file names only, not full patch content
- No merge or rebase operations yet
- Synchronous API only (no Promise support)
- Filter-repo creates bare repositories (can be cloned to get working trees)

## Future Enhancements

Potential additions for future versions:

- Remote operations (fetch, push, pull)
- Merge and rebase support
- Full diff/patch generation
- Progress callbacks for long operations
- Authentication handling for remote operations
- Submodule support
- Promise-based async API

## License

This project is provided as-is for educational and automation purposes.

## References

- [goja](https://github.com/dop251/goja) - ECMAScript 5.1+ implementation in Go
- [go-git](https://github.com/go-git/go-git) - Git implementation in Go
- [isomorphic-git](https://isomorphic-git.org/) - JavaScript Git implementation (API inspiration)
- [simple-git](https://github.com/steveukx/git-js) - Git CLI wrapper for Node.js (API inspiration)
