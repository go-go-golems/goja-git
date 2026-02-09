---
Title: Diary
Ticket: GOJA-001
Status: active
Topics:
    - git
    - goja
    - javascript
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - /home/ubuntu/goja-git/gitmodule.go:Core Git wrapper implementation
    - /home/ubuntu/goja-git/main.go:CLI runner for JS scripts
    - /home/ubuntu/goja-git/scripts/01-init-and-commit.js:Example script for init and commit
    - /home/ubuntu/goja-git/README.md:Comprehensive API documentation
ExternalSources: []
Summary: "Implementation diary for goja Git wrapper module, documenting the discovery of goja's ExportTo quirk and the solutions applied"
LastUpdated: 2026-02-09T08:35:00-05:00
WhatFor: "Track implementation progress and document critical learnings about goja's type system"
WhenToUse: "Reference when working with goja or implementing similar JavaScript wrappers in Go"
---

# Diary

## Goal

Document the implementation of a goja-based Git wrapper module that exposes Git operations to JavaScript, including the challenges encountered with goja's type system and the solutions applied.

## Step 1: Research and Planning

I began by reviewing the user's research materials on implementing Git wrappers for goja. The research included detailed ChatGPT conversations about API design patterns, comparing approaches from `isomorphic-git`, `NodeGit`, and `simple-git`. The key insights were:

- Use a hybrid approach: porcelain methods for common operations, plumbing for advanced use
- Accept single options objects (following isomorphic-git pattern) for easy evolution
- Expose errors as JavaScript exceptions via goja's panic mechanism

**Commit (code):** N/A — planning phase

### What I did

- Read the attached research document (`pasted_content.txt`) covering API design patterns
- Reviewed the diary and git-commit-instructions skills
- Set up Go 1.25.5 toolchain from the provided tarball
- Initialized the project directory and docmgr workspace
- Created ticket GOJA-001 for tracking

### Why

The research provided critical guidance on how to structure the JavaScript API to feel natural while working within goja's constraints. Understanding the differences between `isomorphic-git` (pure JS), `NodeGit` (libgit2 bindings), and `simple-git` (CLI wrapper) helped inform the design decisions.

### What worked

- The research was comprehensive and directly applicable
- docmgr initialization created a clean workspace structure
- Go toolchain installation from tarball was straightforward

### What didn't work

- Initial attempt to use `--project-name` flag with docmgr failed (flag doesn't exist)
- Had to learn docmgr's actual command structure through error messages

### What I learned

- docmgr uses `--root` and `--seed-vocabulary` flags for initialization
- Ticket creation uses `ticket create-ticket --ticket ID --title "..." --topics ...`
- The research emphasized that goja's `ExportTo` function has specific requirements for struct field mapping

### What was tricky to build

Nothing particularly tricky at this stage—mostly setup and planning.

### What warrants a second pair of eyes

The docmgr workflow setup to ensure it follows best practices.

### What should be done in the future

Consider creating a template or script for initializing projects with docmgr to avoid flag discovery through trial and error.

### Code review instructions

N/A for planning phase.

### Technical details

- Go version: 1.25.5
- docmgr initialized at `/home/ubuntu/goja-git/ttmp`
- Ticket: GOJA-001

---

## Step 2: Initial Implementation and Critical Discovery

I implemented the core `gitmodule.go` with the Git wrapper and `main.go` for the CLI runner. The initial implementation followed the research recommendations closely, using option structs with `json` tags.

**Commit (code):** N/A — initial implementation, not yet committed

### What I did

- Created `gitmodule.go` with:
  - `GitModule` for top-level git object
  - `RepoHandle` for repository instances
  - Option structs for all operations (Open, Init, Add, Commit, Log, etc.)
  - Methods for porcelain operations (status, add, commit, log, checkout)
  - Methods for branch and tag management
  - Plumbing methods (refs.resolve)
- Created `main.go` with:
  - Script file loader
  - goja runtime initialization
  - Console.log implementation
  - JSON.stringify helper
- Installed dependencies: `goja` and `go-git/v5`

### Why

The modular structure separates concerns cleanly: `GitModule` handles the JavaScript interface, `RepoHandle` manages repository state, and option structs provide type-safe parameter passing.

### What worked

- Go module initialization and dependency installation succeeded
- Code compiled successfully
- The overall structure followed the research recommendations

### What didn't work

**Critical issue discovered during testing:** goja's `ExportTo` function does NOT respect `json` struct tags. It requires exact case-sensitive field name matching between JavaScript object properties and Go struct fields.

Initial code used:
```go
type OpenOptions struct {
    Dir string `json:"dir"`  // This tag is IGNORED by goja
}
```

JavaScript code:
```javascript
git.open({ dir: "/path" })  // This FAILED
```

The `json:"dir"` tag had no effect—goja's `ExportTo` only matches field names directly. So `{ dir: "..." }` would not map to `Dir string`, but `{ Dir: "..." }` would.

### What I learned

**Critical discovery:** goja's `ExportTo` performs exact case-sensitive field name matching, completely ignoring struct tags like `json:`. This is fundamentally different from Go's standard `json.Unmarshal`.

To verify this, I created a test program:
```go
type TestOpts1 struct {
    Dir string `json:"dir"`
}
type TestOpts2 struct {
    Dir string
}
```

Results:
- `{ dir: "/tmp" }` → `Dir` field remained empty for both structs
- `{ Dir: "/tmp" }` → `Dir` field populated correctly

This meant the entire API needed to use capitalized property names in JavaScript.

### What was tricky to build

**The goja type system quirk** was the trickiest part. The research materials and examples online often show `json` tags, but these don't work with goja's `ExportTo`. This is a subtle but critical difference that's not well-documented.

The debugging process involved:
1. Seeing "dir is required" errors despite passing `{ dir: "..." }`
2. Adding debug logging to inspect what goja was receiving
3. Creating minimal test cases to isolate the issue
4. Discovering the case-sensitivity requirement

### What warrants a second pair of eyes

- The decision to use capitalized property names in the JavaScript API (e.g., `{ Dir: "..." }` instead of `{ dir: "..." }`)
- Whether there's a better way to handle this (custom unmarshaling? wrapper functions?)
- The error messages—should they guide users about capitalization?

### What should be done in the future

- Document this goja quirk prominently in any goja-based project
- Consider creating a helper that allows both lowercase and capitalized property names
- Investigate if newer versions of goja support struct tags

### Code review instructions

**Start here:**
- `/home/ubuntu/goja-git/gitmodule.go` - Core wrapper implementation
  - Look at option struct definitions (lines 18-80)
  - Review the `mustExport` functions (lines 563-583)
  - Check error handling in `Open` and `Init` methods

**How to validate:**
```bash
cd /home/ubuntu/goja-git
go build -o goja-git .

# Test with capitalized properties (should work)
cat > test.js << 'EOF'
const repo = git.open({ Dir: "/home/ubuntu/goja-git/test-repo" });
console.log("Opened:", repo.dir);
EOF
./goja-git test.js

# Test with lowercase properties (will fail)
cat > test.js << 'EOF'
const repo = git.open({ dir: "/home/ubuntu/goja-git/test-repo" });
EOF
./goja-git test.js  # Should error: "Dir is required"
```

### Technical details

**goja ExportTo behavior:**
```go
// This is what goja does internally (simplified):
func ExportTo(jsValue Value, goPtr interface{}) error {
    // Get the Go struct's field names via reflection
    // For each field, look for a JS property with EXACT same name
    // Case-sensitive matching, no tag support
}
```

**Workaround applied:**
- Changed all option structs to use exported field names
- Updated all JavaScript examples to use capitalized properties
- Added clear documentation in README about this requirement

---

## Step 3: JavaScript Script Development

After fixing the type system issues, I created five demonstration scripts showing different Git operations.

**Commit (code):** 10e3ef89ff11301ca554982e606f0ab8903cf673 — "feat: implement goja git wrapper module"

### What I did

- Created `scripts/01-init-and-commit.js` - Repository initialization and first commit
- Created `scripts/02-branching.js` - Branch creation and checkout
- Created `scripts/03-log-and-history.js` - Commit history and reference resolution
- Created `scripts/04-tags-and-diff.js` - Tag management and diff operations
- Created `scripts/05-complete-workflow.js` - Complete workflow demonstration
- Updated all scripts to use capitalized property names
- Created comprehensive README.md with API documentation
- Added .gitignore for build artifacts

### Why

The scripts serve multiple purposes:
1. Demonstrate the API in action
2. Provide copy-paste examples for users
3. Act as integration tests
4. Show a realistic Git workflow

### What worked

All five scripts executed successfully:

**Script 1 output:**
```
=== Script 1: Initialize Repository and Make Commits ===
Repository initialized
Repository at: /home/ubuntu/goja-git/test-repo
Initial status:
  Files: [map[path:README.md ...] ...]
Files added to staging area
First commit created: bb605f00cbff421d633415759e9ebe7a89ad0564
Current branch: main
All branches: [main]
```

**Script 2 output:**
```
Created new branch: feature-test
Checked out branch: feature-test
Current branch is now: feature-test
Switched back to main branch
```

**Script 3 output:**
```
Commit history (last 10 commits):
[map[author:map[email:test@example.com name:Test User ...] ...]]
HEAD resolves to: bb605f00cbff421d633415759e9ebe7a89ad0564
```

**Script 4 output:**
```
Created tag v1.0.0 at: e235847f16ff132b9e9506f475b19da16e21c664
Changed files: [map[from:README.md to:README.md] map[from: to:feature.txt]]
```

**Script 5 output:**
```
Total branches: 2
Total tags: 1
This workflow demonstrated:
  ✓ Status checking
  ✓ Branch inspection
  ✓ Commit history
  ✓ Tag management
  ✓ Reference resolution
```

### What didn't work

- Initially forgot to remove test files (`test-*.go`, `test-*.js`) which caused build errors
- Had to clean these up before the final build

### What I learned

- The iterative test-as-you-go methodology worked extremely well
- Building small, focused scripts is better than one large script
- Error handling with try/catch in JavaScript makes the API user-friendly

### What was tricky to build

**JSON.stringify implementation** in Go was tricky because goja doesn't provide it by default. I had to implement a recursive stringify function that handles:
- Primitives (null, bool, string, number)
- Objects (maps)
- Arrays
- Indentation for pretty-printing

The implementation handles nested structures correctly but doesn't support all JSON.stringify features (like replacer functions or circular reference detection).

### What warrants a second pair of eyes

- The JSON.stringify implementation—it's a simplified version that may not handle all edge cases
- Error messages in the scripts—are they helpful enough for users?
- The script organization—should there be more examples?

### What should be done in the future

- Add more comprehensive error handling examples
- Create scripts for edge cases (empty repos, detached HEAD, etc.)
- Add scripts showing how to handle errors gracefully
- Consider adding a script that demonstrates the full lifecycle of a feature branch

### Code review instructions

**Start here:**
- `/home/ubuntu/goja-git/scripts/01-init-and-commit.js` - Basic operations
- `/home/ubuntu/goja-git/scripts/05-complete-workflow.js` - Complete example

**How to validate:**
```bash
cd /home/ubuntu/goja-git

# Run all scripts in sequence
for script in scripts/0*.js; do
  echo "=== Running $script ==="
  ./goja-git "$script"
  echo ""
done

# Check the test repository state
cd test-repo
git log --oneline --all --graph
git branch -a
git tag
```

### Technical details

**Test repository structure after all scripts:**
```
test-repo/
├── .git/
├── README.md (modified)
├── main.go
├── test.js
└── feature.txt (new)

Branches: main, feature-test
Tags: v1.0.0
Commits: 2
```

**Script execution order:**
1. `01-init-and-commit.js` - Creates initial commit
2. `02-branching.js` - Creates feature-test branch
3. `03-log-and-history.js` - Views history (read-only)
4. `04-tags-and-diff.js` - Creates v1.0.0 tag, shows diff
5. `05-complete-workflow.js` - Summary view (read-only)

---

## Step 4: Documentation and Commit

I created comprehensive documentation and committed everything following git-commit-instructions best practices.

**Commit (code):** 10e3ef89ff11301ca554982e606f0ab8903cf673 — "feat: implement goja git wrapper module"

### What I did

- Created comprehensive README.md covering:
  - Project overview and features
  - Installation and build instructions
  - Complete JavaScript API documentation
  - Example usage for all operations
  - API design notes (especially the capitalization requirement)
  - Architecture overview
  - Limitations and future enhancements
- Created .gitignore to exclude build artifacts and test files
- Reviewed git status before committing
- Staged only production files (excluded test-repo and ttmp)
- Created detailed commit message following conventional commits format
- Recorded commit hash: 10e3ef89ff11301ca554982e606f0ab8903cf673

### Why

Good documentation is critical for a library/wrapper project. Users need to understand:
- How to use the API
- Why certain design decisions were made (capitalization)
- What's possible and what's not (limitations)
- How to extend it (architecture)

### What worked

- The README structure flows logically from overview → usage → API → internals
- The capitalization requirement is prominently documented
- Examples are copy-paste ready
- git-commit-instructions workflow kept the commit clean and focused

### What didn't work

Nothing significant—documentation writing was straightforward.

### What I learned

- Documenting the "why" behind design decisions (like capitalization) is as important as the "how"
- Including limitations upfront sets proper expectations
- The git-commit-instructions workflow (status → diff → stage → commit) prevents accidents

### What was tricky to build

**Balancing detail vs. readability** in the README. Too much detail overwhelms users; too little leaves them confused. I tried to:
- Put essential info up front (features, installation, basic usage)
- Provide comprehensive API reference in the middle
- Include advanced topics (architecture, limitations) at the end

### What warrants a second pair of eyes

- README clarity—is the capitalization requirement explained well enough?
- API documentation completeness—are all methods documented?
- Example quality—are they realistic and helpful?

### What should be done in the future

- Add a "Troubleshooting" section to the README
- Create API reference documentation in a separate file
- Add more code comments in gitmodule.go
- Consider adding godoc comments for Go API documentation

### Code review instructions

**Start here:**
- `/home/ubuntu/goja-git/README.md` - Read through for clarity and completeness
- Check that all methods in `gitmodule.go` are documented in README

**How to validate:**
```bash
# Verify commit follows best practices
cd /home/ubuntu/goja-git
git show --stat 10e3ef8
git log --oneline -1

# Check that no build artifacts or test files were committed
git ls-files | grep -E '(test-|\.exe|\.bin)'  # Should be empty

# Verify .gitignore works
touch test-debug.js goja-git.exe
git status --porcelain  # Should show nothing
```

### Technical details

**Commit structure:**
```
feat: implement goja git wrapper module

- Add gitmodule.go with comprehensive Git operations API
- Add main.go CLI runner for executing JS scripts
- Create 5 example scripts demonstrating different Git operations
- Add README with API documentation and usage examples
- Add .gitignore for build artifacts

Files: 12 changed, 1517 insertions(+)
```

**Files committed:**
- `.gitignore` - Build artifacts exclusion
- `README.md` - Comprehensive documentation
- `gitmodule.go` - Core Git wrapper (583 lines)
- `main.go` - CLI runner (137 lines)
- `go.mod`, `go.sum` - Go module files
- `scripts/*.js` - 6 example scripts

**Files excluded (correctly):**
- `test-repo/` - Test repository (not part of the library)
- `ttmp/` - docmgr workspace (project management, not code)
- `goja-git` - Binary (build artifact)
- Any `test-*.go` or `test-*.js` files

---

## Summary

Successfully implemented a complete goja Git wrapper module with the following achievements:

**Technical accomplishments:**
- ✅ Full Git operations API (init, open, status, add, commit, log, branch, tag, checkout, diff)
- ✅ Clean JavaScript interface following industry patterns (isomorphic-git/simple-git)
- ✅ Comprehensive error handling via JavaScript exceptions
- ✅ 5 working example scripts demonstrating all major operations
- ✅ Complete documentation with API reference and usage examples

**Key learnings:**
- 🔍 goja's `ExportTo` ignores struct tags and requires exact case-sensitive field name matching
- 📝 Capitalized property names in JavaScript (e.g., `{ Dir: "..." }`) are required for goja
- 🧪 Iterative testing with small scripts is highly effective for API development
- 📚 Documenting "why" behind design decisions is as important as "how"

**Challenges overcome:**
- 🐛 Discovered and worked around goja's type system quirk with struct field mapping
- 🔧 Implemented custom JSON.stringify for goja runtime
- 📖 Created comprehensive documentation explaining the capitalization requirement

**What's next:**
- Consider adding remote operations (fetch, push, pull)
- Implement full diff/patch generation
- Add merge and rebase support
- Explore Promise-based async API
- Create helper functions for lowercase property name support

This implementation provides a solid foundation for Git automation in JavaScript via goja, with clear documentation of both capabilities and limitations.
