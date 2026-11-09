# Agents guide

mdembed embeds the contents of code files in Markdown. `README.md`
holds the syntax and the examples.

## Writing

Write every word in ASD-STE100 Simplified Technical English (STE):
Markdown docs, code comments, commit messages, change descriptions,
and replies in an agent conversation. One idea per sentence, active
voice with the actor named, one word for one meaning, and nothing that
carries nothing.
See <https://en.wikipedia.org/wiki/Simplified_Technical_English>.

## Layout

One `main` package at the repo root. `main.go` holds the command, and
`main_test.go` runs it against `examples/`.

## Checks

The root `Checkfile` lists the checks CI runs on every push. Run the
same things before a commit:

```sh
go tool goimports -local "$(go list -m)" -w .
go vet ./...
go test ./...
go tool deadcode -test ./...
git ls-files -z '*.go' | xargs -0 gopls check -severity=hint
```

Pin every tool a check runs. Declare it with
`go get -tool <pkg>@<version>`, so `go.mod` and `go.sum` hold its
version and checksum, and run it as `go tool <name>`. Do not run
`go run pkg@latest` or `go install pkg@latest` in a `Checkfile`, a
script, or CI. A moving version makes the same commit pass one week
and fail the next, and it runs new tool code nobody reviewed. A pinned
tool changes in a change, where the diff shows the update and CI runs
it.

Keep the `tool` lines in `go.mod`, because no module requires mdembed.
A `go get -tool` can raise a requirement, and the build then links the
newer version. Keep the raise only when the tests pass and nothing
holds the old version on purpose. Otherwise, and in a library, put the
tools in a `tools.mod` and run them as
`go tool -modfile=tools.mod <name>`. The hml and uni repos show that
form.

Pin the tool, not its data. The `weekly_govulncheck` check reads the
live vulnerability database on every run, so it still finds advisories
published after the commit. Never vendor or cache that data.

## Changes

Work happens on a sockeye change. `createtree` allocates one and cuts a
worktree, and `soc edit` sets its title and description. Write them
before the work. Push with `soc push --wait`, then merge with
`mergetree`.

## Commits

- Subject: `scope: imperative summary`, at most 72 characters, lowercase
  except proper nouns, no trailing period.
- The scope is what the change acts on: `embed:`, `lang:`, `test:`. Not
  `mdembed:` or `md:`, because every commit here is mdembed.
- Each repo also uses these scopes: `security:` for a fix to a security
  finding, `deps:` for a dependency or the Go version, `doc:` for the
  docs, `todo:` for a plan, and `ci:` for the Checkfile and its checks.
- A change that acts on more than one area takes the scope whose
  behavior changed. Write two scopes, with a comma between them, only
  when each one changed: `embed, lang:`.
- The `merge:` line in the root `Checkfile` checks the change title at
  merge. It caps the text after the scope at 64 characters.
- Include _why_, not just _what_. Hard-wrap the body at 72.
- Write the subject for a teammate who has not read the code. Name the
  language, the flag, or the output, not a term the change makes up.
- Write the body as a change description. sockeye squashes a change into
  one commit whose message is the change's title and description. The
  first push of a branch with one commit copies the commit message to
  the change.
- Sign your work with a `Co-Authored-By` trailer at the end of the
  message, after a blank line. A trailer on the line after the subject
  becomes part of the subject.
