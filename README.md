# Dependabot bundler

Bundler will gather all PRs which were created by `app/dependabot` user. Then, it will apply `go get -u` using the
modules in the prs that it found. It will do that instead of using git magic to combine the prs to avoid the following
problems:

- merge conflicts
- dependencies getting out of sync ( something updating to x while the next downgrades it to y or vica-versa )
- dependency chain conflicts

Once all updates have been applied, it will create a single commit and a PR.

It doesn't attempt to merge PRs causing various merge conflicts. It will basically just do what dependabot would do
but apply it separately as a composite update.

Bundler only ever commits `go.mod` and `go.sum` files. It never stages any other changes.

Example running every Friday:

```yaml
name: Dependabot Bundler

on:
  schedule:
    - 0 0 * * 5 # every Friday at 00:00

jobs:
  bundler:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v2
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.18.x
      - name: Cache go-build and mod
        uses: actions/cache@v2
        with:
          path: |
            ~/.cache/go-build/
            ~/go/pkg/mod/
          key: go-${{ hashFiles('go.sum') }}
          restore-keys: |
            go-          
      - name: Install Dependabot Bundler
        run: |
          go install github.com/Skarlso/dependabot-bundler@v0.0.3
      - name: Run Dependabot Bundler
        run: |
          dependabot-bundler --token ${{ secrets.GITHUB_TOKEN }} --repo test --owner Skarlso
```

If everything goes well, it should result in a PR like this:

![pr1](dummy_sample.png)

This is an actual PR located [here](https://github.com/weaveworks/eksctl/pull/5175) which was created with dependabot-bundler and merged.

![pr2](merged_sample.png)

Dependabot can apply labels to the created PR such as:

```yaml
      - name: Run Dependabot Bundler
        run: |
          dependabot-bundler --token ${{ secrets.GITHUB_TOKEN }} --repo test --owner Skarlso --labels bug,duplicate
```

Which will result in a PR like this:

![pr3](pr_with_labels.png)

## Updating GitHub Actions

Dependabot Bundler is now able to bundle GitHub actions updates as well.

If there are PRs which update the version of GitHub actions, bundler will now take those updates as well
and apply them to the created PR.

![pr4](pr_with_actions.png)

## In Progress features

- [x] define custom labels on the created PR
- [-] define custom description -> decided not to add this unless there is need for it.
- [x] define custom title
