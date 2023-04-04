# Make a release

1. Make a PR onto the main branch that updates the version in the changelog:  
   Insert version (e.g. `## 1.2.3`) inbetween `## Unreleased` and the included changelog entries.  
   Update the version in `internal/version/version.go`.
2. Get the PR merged into the main branch
3. Once the main branch has had its CI steps run, create a tag on that commit and push it:  
   `git tag -a v1.2.3 -m "Release v1.2.3" && git push origin v1.2.3`
