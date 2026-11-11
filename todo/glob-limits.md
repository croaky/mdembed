# Remove brace globs and refuse symlinked directories

Finding: `find_01XCdNebtdQjGEzxjxaDoW7i` (low, CWE-407), "Glob pattern
expansion runs unbounded before the new embed limits apply".

## Threat model

The repo has no `todo/security/README.md`. This plan uses these facts:

- mdembed is a local command. The user pipes their own Markdown
  (`tmp.md`) through it into an LLM tool. The README shows this use.
- The `Checkfile` runs Go checks only. No CI job here runs mdembed on
  a contribution.
- `os.Root` keeps each read in the working directory. A symlink cannot
  move a read outside it.

The finding names a CI job that runs mdembed on pull requests. No such
job exists. The real attacker is the author of a repo that the user
clones and then embeds with a glob, for example `**/*.md`. That author
can put a brace pattern in a nested `.md` file.

- Consequence: the local mdembed process spins on CPU or grows its
  heap. The user stops it with Ctrl-C. No data leaks or changes.
- Boundary: none. The process runs as the user, on the user's input,
  and reads only in the working directory.

## Part 1: brace expansion (CPU and memory)

`doublestar.Glob` expands each `{a,b}` group before it matches
(`globAlts` in doublestar v4.10.0 `glob.go`). n groups of k
alternatives give k^n sub-patterns. The `maxEmbedFiles` check in
`embedPattern` (`main.go:213-221`) runs only after `Glob` returns.

Options:

1. Remove brace alternation. Refuse a pattern that holds `{` with an
   error. The README and the examples do not use braces, so no
   documented feature goes away. Cost: about 3 lines and one test. No
   change to speed or CI time. Without braces, `Glob` work is bounded
   by the size of the tree.
2. Accept the risk, with the attacker, consequence, and boundary
   above. Cost: none. The README then must not claim a bound on work.
3. Full fix to the finding criteria: a work budget shared across the
   run that also counts brace expansions and `**` walks. doublestar
   gives no hook inside `globAlts`, so this needs a fork or a custom
   walker. Cost: about 100 lines or more, a new walker to verify, and
   the loss of a maintained dependency. Not recommended.
4. Replace doublestar with `io/fs.Glob`. This removes a dependency and
   braces, but also removes `**`, which the README documents and
   `examples/input1.md` uses. Not recommended.

Recommendation: option 1.

## Part 2: directory symlinks in a path

`WithNoFollow` does not cover literal segments before the first
metacharacter (doublestar README caveat). `readEmbed`
(`main.go:243-271`) calls `Lstat` on the last path component only. So
`link/f.go` reads through a directory symlink `link`, and one file can
have more than one name. This is false to the README line "It refuses
a symlink". `os.Root` still keeps the read in the working directory.

Options:

1. In `readEmbed`, call `root.Lstat` on each parent directory of the
   match, and refuse a symlink. Cost: about 8 lines and one test, plus
   one `Lstat` for each path segment of each file. That is small next
   to the file read.
2. Change the README to say that a directory symlink in a literal
   segment is followed in the working directory. Cost: one line. The
   cycle check stays keyed by name, and `maxEmbedFiles` still stops a
   cycle that uses an alias name.

Recommendation: option 1. It makes the README claim true with little
code.

## Residual risk to accept

Each of up to 10,000 embed lines can run one full `**` walk of the
tree. Work is lines times tree size, and it is not counted. The
attacker, consequence, and boundary are the same as above. A counter
for this needs the custom walker from part 1, option 3.

## Steps

1. In `embedPattern`, refuse a pattern that holds `{` before the call
   to `Glob`.
2. In `readEmbed`, refuse a match when a parent directory is a
   symlink.
3. Add tests in `main_test.go`: a pattern with many brace groups ends
   at once with an error, and a path through a directory symlink
   gives an error.
4. Update the README "Behavior" section: no brace alternation, and no
   symlink in any path component.
5. Commit as `security: refuse brace globs and symlinked directories`.

## Proposal for the finding

After the change lands, resolve the finding as fixed, with a note that
the uncounted `**` walk per line is an accepted risk.
