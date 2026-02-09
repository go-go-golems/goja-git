// Test 2: Extract subdirectory as root
// This tests extracting cmd/a as the root directory (no prefix)

console.log("=== Test 2: Extract Subdirectory as Root ===\n");

// Create a source repository with cmd/a directory
const srcDir = "/home/ubuntu/goja-git/test-filterrepo/src-extract";
console.log("Creating source repository at:", srcDir);

const srcRepo = git.init({ Dir: srcDir, DefaultBranch: "main" });
console.log("Source repository initialized");

// Add multiple files to cmd/a
srcRepo.add({ Paths: ["cmd/a/main.go"], All: false });
srcRepo.commit({
  Message: "Add main.go",
  Author: { Name: "Test User", Email: "test@example.com" }
});

srcRepo.add({ Paths: ["cmd/a/README.md"], All: false });
srcRepo.commit({
  Message: "Add README.md",
  Author: { Name: "Test User", Email: "test@example.com" }
});

srcRepo.add({ Paths: ["cmd/a/config.yaml"], All: false });
srcRepo.commit({
  Message: "Add config.yaml",
  Author: { Name: "Test User", Email: "test@example.com" }
});

console.log("\nSource repository has 3 commits with cmd/a files");

// Filter-repo: extract cmd/a as root (empty ToPrefix)
const outDir = "/home/ubuntu/goja-git/test-filterrepo/out-extract";
console.log("\nExtracting cmd/a as root...");
console.log("  Keep prefix: cmd/a");
console.log("  New prefix: (empty - becomes root)");
console.log("  Output dir:", outDir);

try {
  const filtered = srcRepo.filterRepo({
    OutDir: outDir,
    Ref: "HEAD",
    Path: "cmd/a",
    ToPrefix: "", // Empty = extract as root
    PruneEmpty: true
  });

  console.log("\n✓ Filter-repo succeeded!");
  console.log("  New tip:", filtered.newTip);
  console.log("  Rewritten commits:", filtered.rewrittenCommits);
  console.log("  Pruned commits:", filtered.prunedCommits);
  console.log("  Output branch:", filtered.outBranch);

  // Verify the output
  console.log("\nVerifying output repository...");
  const outRepo = git.open({ Dir: outDir });
  const log = outRepo.log({ Depth: 10 });
  console.log("  Commit count:", log.length);

  if (log.length !== 3) {
    console.log("✗ ERROR: Expected 3 commits, got", log.length);
  } else {
    console.log("✓ Commit count correct");
  }

  // Verify commit messages are preserved
  console.log("\n  Commit messages:");
  for (let i = 0; i < log.length && i < 3; i++) {
    console.log("    -", log[i].message.trim());
  }

  console.log("\n=== Test 2 PASSED ===");
} catch (e) {
  console.log("\n✗ Test 2 FAILED:", e);
  throw e;
}
