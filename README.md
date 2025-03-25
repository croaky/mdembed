# mdembed

`mdembed` is a command line tool to embed programming file contents in Markdown.

## Install

```sh
go install github.com/croaky/mdembed@latest
```

For that to work, Go must be installed and
[`$(go env GOPATH)/bin`](https://go.dev/wiki/SettingGOPATH)
must be on your `$PATH`:

```sh
export PATH="$HOME/go/bin:$PATH"
```

## Use

In `input1.md`:

````md
# Examples

Embed a whole file:

```embed
dir1/f1.rb
```

Embed multiple whole files:

```embed
dir1/f1.rb
dir1/f2.go
```

Embed specific lines in a file:

```embed
dir1/subdir1/f3.js log
```

Embed multiple whole files and multiple blocks within files:

```embed
dir1/f1.rb
dir1/f2.go log
dir2/f4.css h1
dir2/subdir2/f5.html h1
dir2/subdir2/f5.html ul
f6.sql users
```

Embed files using globs and blocks:

```embed
**/*.rb
**/*.js log
```

Embed Markdown files and their embeds recursively:

```embed
input2.md
```
````

In `dir1/f1.rb`:

```rb
puts "hi"
```

In `dir1/f2.go`:

```go
package main

import "fmt"

func main() {
    fmt.Println("hi")
}
```

In `dir1/subdir1/f3.js`:

```js
console.log("Not embedded");

// emdo log
console.log("hi");
// emdone log

console.log("Not embedded");
```

In `dir2/subdir2/f5.html`:

```html
<!doctype html>
<html>
  <head>
    <title>Not embedded</title>
  </head>
  <body>
    <!-- emdo h1 -->
    <h1>h1</h1>
    <!-- emdone h1 -->
    <p>not embedded</p>
    <!-- emdo ul -->
    <ul>
      <li>1</li>
      <li>2</li>
    </ul>
    <!-- emdone ul -->
  </body>
</html>
```

In `input2.md`:

````md
## Input2

Embed from within an embedded Markdown file:

```embed
dir1/f1.rb
```
````

Run:

```bash
cat input.md | mdembed
```

The output will be:

````md
# Examples

Embed a whole file:

```rb
# dir1/f1.rb
puts "hi"
```

Embed multiple whole files:

```rb
# dir1/f1.rb
puts "hi"
```

```go
// dir1/f2.go
package main

import "fmt"

func main() {
	// emdo log
	fmt.Println("hi")
	// emdone log
}
```

Embed specific lines in a file:

```js
// dir1/subdir1/f3.js
console.log("hi");
```

Embed multiple whole files and multiple blocks within files:

```rb
# dir1/f1.rb
puts "hi"
```

```go
// dir1/f2.go
fmt.Println("hi")
```

```css
/* dir2/f4.css */
h1 {
  color: blue;
}
```

```html
<!-- dir2/subdir2/f5.html -->
<h1>h1</h1>
```

```html
<!-- dir2/subdir2/f5.html -->
<ul>
  <li>1</li>
  <li>2</li>
</ul>
```

```sql
-- f6.sql
SELECT
  *
FROM
  users;
```

Embed files using globs and blocks:

```rb
# dir1/f1.rb
puts "hi"
```

```js
// dir1/subdir1/f3.js
console.log("hi");
```

Embed Markdown files and their embeds recursively:

## Input2

Embed from within an embedded Markdown file:

```rb
# dir1/f1.rb
puts "hi"
```
````

## Behavior

`mdembed` embeds code blocks in the output Markdown,
removing surrounding whitespace.
The file extension is used as the code fence attribute.

It parses exact file paths or file glob patterns.

If `emdo` and `emdone` magic comments were used, it will only embed the code
block wrapped by the magic comments.

It is aware of code comment styles for Ada, Assembly, Awk, Bash, C, Clojure,
COBOL, C++, C#, CSS, CSV, D, Dart, Elm, Erlang, Elixir, Fortran, F#, Gleam, Go,
Haml, Haskell, HTML, Java, Julia, JavaScript, JSON, JSX, Kotlin, Lisp, Logo,
Lua, MATLAB, OCaml, Objective-C, Mojo, Nim, Pascal, PHP, Perl, Prolog, Python,
R, Ruby, Rust, Scala, Scheme, Sass, Shell, Solidity, SQL, Swift, Tcl,
TypeScript, TSX, VBScript, Visual Basic, Wolfram, YAML, and Zig.

If you reference another Markdown file, `mdembed` will embed its contents
directly, recursively embedding its code blocks.

## So what?

I wanted the following workflow in Vim:

1. Open `tmp.md` in my project.
2. Write a prompt for an LLM (Large Language Model),
   referencing other files, or subsets of files, in my project.
3. Hit a key combo (`<Leader>+r`) to send the contents to an LLM
   and render the LLM's response in a vertical split.

`mdembed` handles the Markdown parsing steps.
[mods](https://github.com/charmbracelet/mods) handles the LLM steps:

```bash
go install github.com/charmbracelet/mods@latest
```

So, my Vim config runs the following Unix pipeline in a vertical split:

```bash
cat example.md | mdembed | mods
```

## License

MIT
