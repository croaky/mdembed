# Example

Embed a whole file:

```rb
puts "f1.rb"
```

Embed multiple whole files:

```rb
puts "f1.rb"
puts "f2.rb"
```

Embed subsets of a file using `beginembed` and `endembed` magic comments:

```rb
puts "f3.rb"
```

Embed multiple whole files and multiple subsets of files:

```rb
puts "f1.rb"
puts "f2.rb"
puts "f3.rb"
puts "f4.rb"
```
