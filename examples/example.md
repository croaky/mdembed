# Example

Embed a whole file:

```embed
f1.rb
```

Embed multiple whole files:

```embed
f1.rb
f2.go
```

Embed subsets of a file using `beginembed` and `endembed` magic comments:

```embed
f3.js subset
```

Embed multiple whole files and multiple subsets of files:

```embed
f1.rb
f4.css subset
f5.html subset
```
