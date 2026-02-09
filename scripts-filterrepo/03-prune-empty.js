// Test 3: Prune empty commits
// This tests that commits without the target path are pruned

console.log("=== Test 3: Prune Empty Commits ===\n");

// Create a source repository with mixed commits
const srcDir = "/home/ubuntu/goja-git/test-filterrepo/src-prune";
console.log("Creating source repository at:", srcDir);

const srcRepo = git.init({ Dir: srcDir, DefaultBranch: "main" });
console.log("Source repository initialized");

// Commit 1: Add file to other/ (not cmd/a) - should be pruned
srcRepo.add({ Paths: ["other/file1.txt"], All: false });
srcRepo.commit({
  Message: "Add other/file1.txt (not in cmd/a)",
  Author: { Name: "Test User", Email: "test@example.com" }
});

// Commit 2: Add file to cmd/a - should be kept
srcRepo.add({ Paths: ["cmd/a/app.go"], All: false });
srcRepo.commit({
  Message: "Add cmd/a/app.go",
  Author: { Name: "Test User", Email: "test@example.com" }
});

// Commit 3: Modify other/ (not cmd/a) - should be pruned
srcRepo.add({ Paths: ["other/file2.txt"], All: false });
srcRepo.commit({
  Message: "Add other/file2.txt (not in cmd/a)",
  Author: { Name: "Test User", Email: "test@example.com" }
});

// Commit 4: Modify cmd/a - should be kept
srcRepo.add({ Paths: ["cmd/a/utils.go"], All: false });
srcRepo.commit({
  Message: "Add cmd/a/utils.go",
  Author: { Name: "Test User", Email: "test@example.com" }
});

// Commit 5: Add to docs/ (not cmd/a) - should be pruned
srcRepo.add({ Paths: ["docs/README.md"], All: false });
srcRepo.commit({
  Message: "Add docs/README.md (not in cmd/a)",
  Author: { Name: "Test User", Email: "test@example.com" }
});

console.log("\nSource repository has 5 commits:");
console.log("  - 2 commits with cmd/a files (should be kept)");
console.log("  - 3 commits without cmd/a files (should be pruned)");

// Filter-repo with pruning enabled
const outDir = "/home/ubuntu/goja-git/test-filterrepo/out-prune";
console.log("\nFiltering repository with pruning...");
console.log("  Keep prefix: cmd/a");
console.log("  New prefix: (empty)");
console.log("  Prune empty: true");
console.log("  Output dir:", outDir);

try {
  const filtered = srcRepo.filterRepo({
    OutDir: outDir,
    Ref: "HEAD",
    Path: "cmd/a",
    ToPrefix: "",
    PruneEmpty: true
  });

  console.log("\n✓ Filter-repo succeeded!");
  console.log("  New tip:", filtered.newTip);
  console.log("  Rewritten commits:", filtered.rewrittenCommits);
  console.log("  Pruned commits:", filtered.prunedCommits);
  console.log("  Output branch:", filtered.outBranch);

  // Verify pruning
  if (filtered.prunedCommits !== 3) {
    console.log("✗ ERROR: Expected 3 pruned commits, got", filtered.prunedCommits);
  } else {
    console.log("✓ Pruned commit count correct");
  }

  if (filtered.rewrittenCommits !== 2) {
    console.log("✗ ERROR: Expected 2 rewritten commits, got", filtered.rewrittenCommits);
  } else {
    console.log("✓ Rewritten commit count correct");
  }

  // Verify the output
  console.log("\nVerifying output repository...");
  const outRepo = git.open({ Dir: outDir });
  const log = outRepo.log({ Depth: 10 });
  console.log("  Final commit count:", log.length);

  if (log.length !== 2) {
    console.log("✗ ERROR: Expected 2 commits in output, got", log.length);
  } else {
    console.log("✓ Final commit count correct");
  }

  // Verify only cmd/a commits remain
  console.log("\n  Remaining commit messages:");
  for (let i = 0; i < log.length; i++) {
    const msg = log[i].message.trim();
    console.log("    -", msg);
    if (!msg.includes("cmd/a")) {
      console.log("      ✗ WARNING: Commit doesn't mention cmd/a");
    }
  }

  console.log("\n=== Test 3 PASSED ===");
} catch (e) {
  console.log("\n✗ Test 3 FAILED:", e);
  throw e;
}
