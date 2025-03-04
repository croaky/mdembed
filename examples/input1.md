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
