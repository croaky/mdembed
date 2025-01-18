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
