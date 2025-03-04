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
