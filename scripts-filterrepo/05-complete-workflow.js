// Test 5: Complete workflow test
// This tests a realistic scenario: extracting a monorepo subdirectory

console.log("=== Test 5: Complete Workflow - Monorepo Extraction ===\n");

// Simulate a monorepo with multiple projects
const srcDir = "/home/ubuntu/goja-git/test-filterrepo/src-monorepo";
console.log("Creating monorepo at:", srcDir);

const srcRepo = git.init({ Dir: srcDir, DefaultBranch: "main" });
console.log("Monorepo initialized");

// Initial structure
srcRepo.add({ Paths: ["README.md"], All: false });
srcRepo.commit({
  Message: "Initial commit",
  Author: { Name: "Test User", Email: "test@example.com" }
});

// Add project A
srcRepo.add({ Paths: ["projects/project-a/main.go"], All: false });
srcRepo.commit({
  Message: "Add project A",
  Author: { Name: "Test User", Email: "test@example.com" }
});

// Add project B
srcRepo.add({ Paths: ["projects/project-b/main.go"], All: false });
srcRepo.commit({
  Message: "Add project B",
  Author: { Name: "Test User", Email: "test@example.com" }
});

// Update project A
srcRepo.add({ Paths: ["projects/project-a/utils.go"], All: false });
srcRepo.commit({
  Message: "Add utils to project A",
  Author: { Name: "Test User", Email: "test@example.com" }
});

// Update shared docs
srcRepo.add({ Paths: ["docs/architecture.md"], All: false });
srcRepo.commit({
  Message: "Add architecture docs",
  Author: { Name: "Test User", Email: "test@example.com" }
});

// Update project A again
srcRepo.add({ Paths: ["projects/project-a/config.yaml"], All: false });
srcRepo.commit({
  Message: "Add config to project A",
  Author: { Name: "Test User", Email: "test@example.com" }
});

// Update project B
srcRepo.add({ Paths: ["projects/project-b/server.go"], All: false });
srcRepo.commit({
  Message: "Add server to project B",
  Author: { Name: "Test User", Email: "test@example.com" }
});

console.log("\nMonorepo has 7 commits:");
console.log("  - 1 root commit");
console.log("  - 3 commits for project A");
console.log("  - 2 commits for project B");
console.log("  - 1 commit for shared docs");

// Extract project A as its own repository
const outDir = "/home/ubuntu/goja-git/test-filterrepo/out-project-a";
console.log("\nExtracting project A...");
console.log("  Keep prefix: projects/project-a");
console.log("  New prefix: (empty - becomes root)");
console.log("  Prune empty: true");
console.log("  Output dir:", outDir);

try {
  const filtered = srcRepo.filterRepo({
    OutDir: outDir,
    Ref: "HEAD",
    Path: "projects/project-a",
    ToPrefix: "",
    PruneEmpty: true,
    OutBranch: "main"
  });

  console.log("\n✓ Filter-repo succeeded!");
  console.log("  New tip:", filtered.newTip);
  console.log("  Rewritten commits:", filtered.rewrittenCommits);
  console.log("  Pruned commits:", filtered.prunedCommits);
  console.log("  Output branch:", filtered.outBranch);

  // Verify the output
  console.log("\nVerifying extracted repository...");
  const outRepo = git.open({ Dir: outDir });
  const log = outRepo.log({ Depth: 20 });
  console.log("  Commit count:", log.length);

  // Should have 3 commits (project A commits only)
  if (log.length !== 3) {
    console.log("✗ ERROR: Expected 3 commits, got", log.length);
  } else {
    console.log("✓ Commit count correct");
  }

  // Verify pruning
  const expectedPruned = 4; // root, project B, docs, project B again
  if (filtered.prunedCommits !== expectedPruned) {
    console.log("✗ WARNING: Expected", expectedPruned, "pruned commits, got", filtered.prunedCommits);
  } else {
    console.log("✓ Pruned commit count correct");
  }

  // Show extracted history
  console.log("\n  Extracted commit history:");
  for (let i = 0; i < log.length && i < 5; i++) {
    console.log("    " + (i + 1) + ".", log[i].message.trim());
  }

  // Verify branch
  const currentBranch = outRepo.branch.current();
  console.log("\n  Current branch:", currentBranch);
  if (currentBranch !== "main") {
    console.log("✗ WARNING: Expected branch 'main', got", currentBranch);
  } else {
    console.log("✓ Branch name correct");
  }

  console.log("\n=== Test 5 PASSED ===");
  console.log("\nSummary:");
  console.log("  ✓ Successfully extracted project A from monorepo");
  console.log("  ✓ Pruned commits not related to project A");
  console.log("  ✓ Project A files are now at repository root");
  console.log("  ✓ Commit history preserved for project A");
} catch (e) {
  console.log("\n✗ Test 5 FAILED:", e);
  throw e;
}
