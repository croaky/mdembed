# Refuse a pattern with more than one `**` segment

Finding: `find_019Arw4WLN2QQk3BcZresdj9` (medium, CWE-407), "Repeated
\*\* glob segments cause combinatorial directory walks before embed
limits apply".

## Threat model

The repo has no `todo/security/README.md`. This plan uses the same
facts as the plan that commit 5547c1b closed:

- mdembed is a local command. The user pipes their own Markdown
  (`tmp.md`) through it into an LLM tool. The README shows this use.
- The `Checkfile` runs Go checks only. No CI job runs mdembed on a
  contribution.
- `os.Root` keeps each read in the working directory.

The finding names a CI docs build. No such job exists. The real
attacker is the author of a repo that the user clones and then embeds
with an outer glob such as `**/*.md`. That author can put a pattern
with many `**` segments in a nested `.md` file.

- Consequence: the local mdembed process spins on CPU and grows its
  heap. The user stops it with Ctrl-C. No data leaks or changes.
- Boundary: none. The process runs as the user and reads only in the
  working directory.

## What the finding gets right

Commit 5547c1b accepted "one full `**` walk of the tree per embed
line". That claim is false for more than one `**`. A test on
doublestar v4.10.0, with one file five directories deep:

- `**/**/x.go`: 6 matches, under 1 s.
- 6 `**` segments: 252 matches, 1 s.
- 10 `**` segments: 2002 matches, 6 s.

Each match is the same file. mdembed embeds each copy, so the output
is also wrong, not only slow. A user who types `**/**/x.go` by mistake
gets the file six times. `maxEmbedFiles` stops the output at 10,000
copies, but only after `doublestar.Glob` returns
(`main.go (219-227)`).

## Options

1. Refuse a pattern that holds more than one `**` segment, before the
   call to `Glob` in `embedPattern` (`main.go (199-213)`). The README
   and the examples use one `**` at most (`examples/input1.md:36-37`).
   With one `**`, each match has one split, so the glob gives no
   duplicate, and the walk visits each directory a bounded number of
   times. Cost: about 4 lines and one test. No change to speed or CI
   time. A user loses patterns such as `a/**/b/**/x.go`. No doc shows
   that form.
2. Collapse consecutive `**/` segments to one before the glob. Cost:
   about 3 lines. This keeps `a/**/b/**/x.go`, but interleaved forms
   such as `**/*/**/*/x.go` still grow. It does not meet the fix
   criteria. Not recommended alone.
3. Remove duplicates from the match list after the glob. Cost: about
   5 lines. This fixes the output but not the walk. Not recommended
   alone.
4. Replace `doublestar.Glob` with one `fs.WalkDir` and
   `doublestar.Match` on each path. Cost: about 30 lines, a walk of
   the full tree for each pattern, also for a literal path such as
   `main.go`. That makes the usual case slower. `Match` also
   backtracks on `**`, so the bound must be proven again. Not
   recommended.
5. Accept the risk with the attacker, consequence, and boundary
   above. Cost: none in code. The duplicate output for `**/**` stays.
   Not recommended, because option 1 costs about as little as a
   README note and also fixes wrong output.

Recommendation: option 1.

## Steps

1. In `embedPattern`, count the path segments equal to `**`. Refuse
   the pattern with an error when the count is more than one. Update
   the comment above `embedPattern`.
2. Add a test in `main_test.go`: a nested tree and a pattern of 200
   `**` segments. The run ends at once with an error. Run it from
   stdin and from a nested `.md` file that an outer `**/*.md` embeds.
3. Add a test that `**/x.go` still embeds each match one time.
4. Update the README "Behavior" section: one `**` segment at most in a
   pattern.
5. Commit as `security: refuse globs with more than one ** segment`.
   Delete this plan in that commit.

## Proposal for the finding

After the change lands, resolve the finding as fixed. Note that
mdembed now refuses each pattern with more than one `**` segment, so
the per-line walk is linear in the tree size.
