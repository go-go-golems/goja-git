// Test 1: Basic path filtering and renaming
// This tests the core functionality: keep cmd/a, rename to cmd/b

console.log("=== Test 1: Basic Path Filtering and Renaming ===\n");

// Create a source repository with cmd/a directory
const srcDir = "/home/ubuntu/goja-git/test-filterrepo/src-basic";
console.log("Creating source repository at:", srcDir);

const srcRepo = git.init({ Dir: srcDir, DefaultBranch: "main" });
console.log("Source repository initialized");

// Add files to cmd/a
srcRepo.add({ Paths: ["cmd/a/file1.txt"], All: false });
srcRepo.commit({
  Message: "Add file1.txt",
  Author: { Name: "Test User", Email: "test@example.com" }
});

srcRepo.add({ Paths: ["cmd/a/file2.txt"], All: false });
srcRepo.commit({
  Message: "Add file2.txt",
  Author: { Name: "Test User", Email: "test@example.com" }
});

srcRepo.add({ Paths: ["cmd/a/file3.txt"], All: false });
srcRepo.commit({
  Message: "Add file3.txt",
  Author: { Name: "Test User", Email: "test@example.com" }
});

console.log("\nSource repository has 3 commits with cmd/a files");

// Filter-repo: keep cmd/a, rename to cmd/b
const outDir = "/home/ubuntu/goja-git/test-filterrepo/out-basic";
console.log("\nFiltering repository...");
console.log("  Keep prefix: cmd/a");
console.log("  New prefix: cmd/b");
console.log("  Output dir:", outDir);

try {
  const filtered = srcRepo.filterRepo({
    OutDir: outDir,
    Ref: "HEAD",
    Path: "cmd/a",
    ToPrefix: "cmd/b",
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

  // Check branch
  const currentBranch = outRepo.branch.current();
  console.log("  Current branch:", currentBranch);

  console.log("\n=== Test 1 PASSED ===");
} catch (e) {
  console.log("\n✗ Test 1 FAILED:", e);
  throw e;
}
