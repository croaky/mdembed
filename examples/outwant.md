# Examples

Embed a whole file:

```rb
# f1.rb
puts "hi"
```

Embed multiple whole files:

```rb
# f1.rb
puts "hi"
```

```go
// f2.go
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
// f3.js
console.log("hi");
```

Embed multiple whole files and multiple blocks within files:

```rb
# f1.rb
puts "hi"
```

```go
// f2.go
fmt.Println("hi")
```

```css
/* f4.css */
h1 {
  color: blue;
}
```

```html
<!-- f5.html -->
<h1>h1</h1>
```

```html
<!-- f5.html -->
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

Embed Markdown files and their embeds recursively:

## Input2

Embed from within an embedded Markdown file:

```rb
# f1.rb
puts "hi"
```
