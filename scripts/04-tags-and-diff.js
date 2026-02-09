// Script 4: Tags and diff operations

console.log("=== Script 4: Tags and Diff Operations ===\n");

// Open existing repository
const repo = git.open({
  Dir: "/home/ubuntu/goja-git/test-repo"
});

console.log("Opened repository at:", repo.dir);

// List existing tags
const tags = repo.tag.list();
console.log("\nExisting tags:", JSON.stringify(tags, null, 2));

// Create a new tag
try {
  const tagHash = repo.tag.create({
    Name: "v1.0.0",
    Message: "First release",
    Ref: "HEAD"
  });
  console.log("\nCreated tag v1.0.0 at:", tagHash);
} catch (e) {
  console.log("Error creating tag:", e.message);
  console.log("(This is expected if no commits exist yet)");
}

// List tags after creation
const tagsAfter = repo.tag.list();
console.log("\nTags after creation:", JSON.stringify(tagsAfter, null, 2));

// Try to show diff between commits
try {
  const log = repo.log({ Depth: 2 });
  
  if (log.length >= 2) {
    console.log("\n--- Showing diff between last 2 commits ---");
    const diff = repo.diff({
      From: log[1].oid,
      To: log[0].oid
    });
    console.log("Changed files:", JSON.stringify(diff, null, 2));
  } else {
    console.log("\nNot enough commits to show diff (need at least 2)");
  }
} catch (e) {
  console.log("Error showing diff:", e.message);
}

// Show current repository status
const status = repo.status();
console.log("\nCurrent repository status:");
if (status.length === 0) {
  console.log("  Working tree clean");
} else {
  console.log(JSON.stringify(status, null, 2));
}

console.log("\n=== Script 4 Complete ===");
