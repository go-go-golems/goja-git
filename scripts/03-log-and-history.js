// Script 3: View commit history and logs

console.log("=== Script 3: Commit History and Logs ===\n");

// Open existing repository
const repo = git.open({
  Dir: "/home/ubuntu/goja-git/test-repo"
});

console.log("Opened repository at:", repo.dir);

// Get commit log
try {
  const log = repo.log({
    Ref: "HEAD",
    Depth: 10
  });
  
  console.log("\nCommit history (last 10 commits):");
  console.log(JSON.stringify(log, null, 2));
  
  if (log.length > 0) {
    console.log("\n--- Formatted commit history ---");
    for (let i = 0; i < log.length; i++) {
      const commit = log[i];
      console.log("\nCommit " + (i + 1) + ":");
      console.log("  Hash:", commit.oid);
      console.log("  Author:", commit.author.name, "<" + commit.author.email + ">");
      console.log("  Date:", commit.author.when);
      console.log("  Message:", commit.message.trim());
    }
  }
} catch (e) {
  console.log("Error getting log:", e.message);
  console.log("(This is expected if no commits exist yet)");
}

// Try to resolve HEAD reference
try {
  const headHash = repo.refs.resolve({ Ref: "HEAD" });
  console.log("\nHEAD resolves to:", headHash);
} catch (e) {
  console.log("\nCould not resolve HEAD:", e.message);
}

// Try to resolve a specific branch
try {
  const mainHash = repo.refs.resolve({ Ref: "main" });
  console.log("main branch resolves to:", mainHash);
} catch (e) {
  console.log("Could not resolve main:", e.message);
}

console.log("\n=== Script 3 Complete ===");
