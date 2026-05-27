# goja-git xgoja provider smoke

Builds a generated xgoja binary that imports `github.com/go-go-golems/goja-git/pkg/provider`, selects module `git`, and runs a non-destructive API-shape smoke script.

The runtime profile sets `allowWrite: true` because the provider can create or mutate repositories. The smoke script only checks exported functions.

```bash
make smoke
```
