const git = require("git");
if (typeof git.open !== "function") throw new Error("missing git.open");
if (typeof git.init !== "function") throw new Error("missing git.init");
console.log("goja-git provider ok");
