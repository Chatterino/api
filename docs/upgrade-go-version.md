When a new Go version has been out for a while, we'll want to transition to testing/supporting only the latest 2 versions of Go.

These are the steps you can follow to ensure any place where the Go version is mentioned is kept up to date:

1. Update go matrix in `.github/workflows/lint.yml`
2. Update go matrix in `.github/workflows/build.yml`
3. Update which artifact is downloaded `.github/workflows/build.yml` in the create-release job
4. Bump `./Dockerfile` version
5. Update the branch protection rules to remove the removed go version, and add the newly added go version
