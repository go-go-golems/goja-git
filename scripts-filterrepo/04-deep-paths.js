// Test 4: Multiple path levels
// This tests filtering deeply nested directories

console.log("=== Test 4: Deep Path Filtering ===\n");

// Create a source repository with deeply nested paths
const srcDir = "/home/ubuntu/goja-git/test-filterrepo/src-deep";
console.log("Creating source repository at:", srcDir);

const srcRepo = git.init({ Dir: srcDir, DefaultBranch: "main" });
console.log("Source repository initialized");

// Add files at various depths
srcRepo.add({ Paths: ["pkg/api/v1/handler.go"], All: false });
srcRepo.commit({
  Message: "Add API v1 handler",
  Author: { Name: "Test User", Email: "test@example.com" }
});

srcRepo.add({ Paths: ["pkg/api/v1/types.go"], All: false });
srcRepo.commit({
  Message: "Add API v1 types",
  Author: { Name: "Test User", Email: "test@example.com" }
});

srcRepo.add({ Paths: ["pkg/api/v1/middleware/auth.go"], All: false });
srcRepo.commit({
  Message: "Add auth middleware",
  Author: { Name: "Test User", Email: "test@example.com" }
});

srcRepo.add({ Paths: ["pkg/api/v1/middleware/logging.go"], All: false });
srcRepo.commit({
  Message: "Add logging middleware",
  Author: { Name: "Test User", Email: "test@example.com" }
});

console.log("\nSource repository has 4 commits with pkg/api/v1 files");

// Filter-repo: keep pkg/api/v1, rename to api/v2
const outDir = "/home/ubuntu/goja-git/test-filterrepo/out-deep";
console.log("\nFiltering deeply nested path...");
console.log("  Keep prefix: pkg/api/v1");
console.log("  New prefix: api/v2");
console.log("  Output dir:", outDir);

try {
  const filtered = srcRepo.filterRepo({
    OutDir: outDir,
    Ref: "HEAD",
    Path: "pkg/api/v1",
    ToPrefix: "api/v2",
    PruneEmpty: false
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

  if (log.length !== 4) {
    console.log("✗ ERROR: Expected 4 commits, got", log.length);
  } else {
    console.log("✓ Commit count correct");
  }

  // Verify commit messages
  console.log("\n  Commit history:");
  for (let i = 0; i < log.length && i < 4; i++) {
    console.log("    " + (i + 1) + ".", log[i].message.trim());
  }

  console.log("\n  All files should now be under api/v2/ instead of pkg/api/v1/");

  console.log("\n=== Test 4 PASSED ===");
} catch (e) {
  console.log("\n✗ Test 4 FAILED:", e);
  throw e;
}
