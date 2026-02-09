// Script 5: Complete workflow demonstration
// This script demonstrates a realistic git workflow

console.log("=== Script 5: Complete Git Workflow ===\n");

// Open the repository
const repo = git.open({
  Dir: "/home/ubuntu/goja-git/test-repo"
});

console.log("Working with repository at:", repo.dir);

// Step 1: Check current status
console.log("\n--- Step 1: Check Status ---");
const initialStatus = repo.status();
console.log("Repository status:");
if (initialStatus.length === 0) {
  console.log("  Working tree clean");
} else {
  console.log("  Modified files:", initialStatus.length);
  for (let i = 0; i < initialStatus.length; i++) {
    console.log("    -", initialStatus[i].path, 
                "(staging:", initialStatus[i].staging, 
                "worktree:", initialStatus[i].worktree + ")");
  }
}

// Step 2: Show current branch
console.log("\n--- Step 2: Current Branch ---");
try {
  const currentBranch = repo.branch.current();
  console.log("On branch:", currentBranch);
} catch (e) {
  console.log("Error:", e.message);
}

// Step 3: List all branches
console.log("\n--- Step 3: All Branches ---");
const allBranches = repo.branch.list();
console.log("Branches:", JSON.stringify(allBranches, null, 2));

// Step 4: Show recent commits
console.log("\n--- Step 4: Recent Commits ---");
try {
  const commits = repo.log({ Depth: 5 });
  console.log("Last", commits.length, "commits:");
  for (let i = 0; i < commits.length; i++) {
    const c = commits[i];
    console.log("\n  Commit:", c.oid.substring(0, 8));
    console.log("  Author:", c.author.name);
    console.log("  Date:", c.author.when);
    console.log("  Message:", c.message.trim());
  }
} catch (e) {
  console.log("Error getting commits:", e.message);
}

// Step 5: List tags
console.log("\n--- Step 5: Tags ---");
const tags = repo.tag.list();
if (tags.length === 0) {
  console.log("No tags found");
} else {
  console.log("Tags:", JSON.stringify(tags, null, 2));
}

// Step 6: Resolve references
console.log("\n--- Step 6: Reference Resolution ---");
try {
  const headHash = repo.refs.resolve({ Ref: "HEAD" });
  console.log("HEAD ->", headHash);
} catch (e) {
  console.log("Could not resolve HEAD:", e.message);
}

// Step 7: Summary
console.log("\n--- Summary ---");
console.log("Repository:", repo.dir);
try {
  console.log("Current branch:", repo.branch.current());
} catch (e) {
  console.log("Current branch: (error)");
}
console.log("Total branches:", allBranches.length);
console.log("Total tags:", tags.length);

console.log("\n=== Script 5 Complete ===");
console.log("\nThis workflow demonstrated:");
console.log("  ✓ Status checking");
console.log("  ✓ Branch inspection");
console.log("  ✓ Commit history");
console.log("  ✓ Tag management");
console.log("  ✓ Reference resolution");
