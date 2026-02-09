// Script 2: Branch operations

console.log("=== Script 2: Branch Operations ===\n");

// Open existing repository
const repo = git.open({
  Dir: "/home/ubuntu/goja-git/test-repo"
});

console.log("Opened repository at:", repo.dir);

// Get current branch
try {
  const currentBranch = repo.branch.current();
  console.log("Current branch:", currentBranch);
} catch (e) {
  console.log("Error getting current branch:", e.message);
}

// List existing branches
const branches = repo.branch.list();
console.log("\nExisting branches:", JSON.stringify(branches, null, 2));

// Create a new feature branch
try {
  repo.branch.create({
    Name: "feature-test",
    StartPoint: "HEAD"
  });
  console.log("\nCreated new branch: feature-test");
} catch (e) {
  console.log("Error creating branch:", e.message);
}

// List branches again
const branchesAfter = repo.branch.list();
console.log("\nBranches after creation:", JSON.stringify(branchesAfter, null, 2));

// Checkout the new branch
try {
  repo.checkout({
    Ref: "feature-test",
    Create: false
  });
  console.log("\nChecked out branch: feature-test");
  
  const newCurrent = repo.branch.current();
  console.log("Current branch is now:", newCurrent);
} catch (e) {
  console.log("Error checking out branch:", e.message);
}

// Check status on new branch
const status = repo.status();
console.log("\nStatus on feature-test branch:");
if (status.length === 0) {
  console.log("  Working tree clean");
} else {
  console.log(JSON.stringify(status, null, 2));
}

// Switch back to main
try {
  repo.checkout({
    Ref: "main",
    Create: false
  });
  console.log("\nSwitched back to main branch");
  
  const finalBranch = repo.branch.current();
  console.log("Current branch:", finalBranch);
} catch (e) {
  console.log("Error switching back:", e.message);
}

console.log("\n=== Script 2 Complete ===");
