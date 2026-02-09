const repo = git.open({ Dir: "/home/ubuntu/goja-git/test-repo" });
repo.add({ All: true });
const hash = repo.commit({
  Message: "Add feature.txt and update README",
  Author: { Name: "Test User", Email: "test@example.com" }
});
console.log("Created commit:", hash);
