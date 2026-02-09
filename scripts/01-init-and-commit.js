// Script 1: Initialize a repository and make some commits

console.log("=== Script 1: Initialize Repository and Make Commits ===\n");

// Initialize a new repository (or open if already exists)
let repo;
try {
  repo = git.init({
    Dir: "/home/ubuntu/goja-git/test-repo",
    DefaultBranch: "main"
  });
  console.log("Repository initialized");
} catch (e) {
  console.log("Repository already exists, opening it:", e.message);
  repo = git.open({
    Dir: "/home/ubuntu/goja-git/test-repo"
  });
}

console.log("Repository at:", repo.dir);

// Check initial status
const status1 = repo.status();
console.log("\nInitial status:");
if (status1.length === 0) {
  console.log("  Working tree clean");
} else {
  console.log("  Files:", JSON.stringify(status1, null, 2));
}

// Add all files
try {
  repo.add({ All: true });
  console.log("\nFiles added to staging area");
  
  const status2 = repo.status();
  console.log("Status after add:", JSON.stringify(status2, null, 2));
} catch (e) {
  console.log("Error adding files:", e.message);
}

// Make initial commit
try {
  const commitHash = repo.commit({
    Message: "Initial commit",
    Author: {
      Name: "Test User",
      Email: "test@example.com"
    }
  });
  console.log("\nFirst commit created:", commitHash);
} catch (e) {
  console.log("Could not create commit:", e.message);
}

// Check current branch
try {
  const currentBranch = repo.branch.current();
  console.log("\nCurrent branch:", currentBranch);
} catch (e) {
  console.log("Could not get current branch:", e.message);
}

// List all branches
const branches = repo.branch.list();
console.log("All branches:", JSON.stringify(branches, null, 2));

console.log("\n=== Script 1 Complete ===");
