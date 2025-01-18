# mdembed

`mdembed` is a command line tool to embed file contents in Markdown files
in Unix pipelines.

## Quick Start

Install:

```bash
go install github.com/croaky/mdembed
```

Use:

```bash
cat example.md | mdembed
```

## Detailed Example

See the [example](https://github.com/croaky/mdembed/tree/main/example)
directory in this repo.

The directory contains these files:

```
.
├── example.md
├── f1.rb
├── f2.rb
├── f3.rb
└── f4.rb
```

`f1.rb` contains:

```ruby
puts "f1.rb"
```

`f2.rb` contains:

```ruby
puts "f2.rb"
```

`f3.rb` contains:

```ruby
puts "not embedded"

# beginembed
puts "f3.rb"
# endembed

puts "not embedded"
```

`f4.rb` contains:

```ruby
puts "not embedded"

# beginembed
puts "f4.rb"
# endembed

puts "not embedded"
```

To embed whole files, or subsets of files,
use fenced code blocks with an `embed` attribute.

In `example.md`:

    # Example

    Embed a whole file:

    ```embed
    f1.rb
    ```

    Embed multiple whole files:

    ```embed
    f1.rb
    f2.rb
    ```

    Embed subsets of a file using `beginembed` and `endembed` magic comments:

    ```embed
    f3.rb subset
    ```

    Embed multiple whole files and multiple subsets of files:

    ```embed
    f1.rb
    f2.rb
    f3.rb subset
    f4.rb subset
    ```

Run:

```bash
cat example.md | mdembed
```

Output:

    # Example

    Embed a whole file:

    ```ruby
    puts "f1.rb"
    ```

    Embed multiple whole files:

    ```ruby
    puts "f1.rb"
    puts "f2.rb"
    ```

    Embed subsets of a file using `beginembed` and `endembed` magic comments:

    ```ruby
    puts "f3.rb"
    ```

    Embed multiple whole files and multiple subsets of files:

    ```ruby
    puts "f1.rb"
    puts "f2.rb"
    puts "f3.rb"
    puts "f4.rb"
    ```

The resulting Markdown has replaced the `embed` blocks with
the contents of the files, or subsets of files,
and replaced the attribute for the code fence based on the embedded file's
file type.

## So What?

I wanted the following workflow in Vim:

1. Open `tmp.md` in my project.
2. Write a prompt for an LLM (Large Language Model).
3. Reference other files, or subsets of files, in my project.
4. Hit a key combo (`Space+r` for me) to send all the contents to an LLM.
5. Open a visual split to render the LLM's response.

`mdembed` handles step 3.
I use [mods](https://github.com/charmbracelet/mods) for the LLM steps:

```bash
go install github.com/charmbracelet/mods@latest
```

So, my Unix pipeline is:

```bash
cat example.md | mdembed | mods
```
