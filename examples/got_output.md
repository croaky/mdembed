# Example

Embed a whole file:

```rb
# f1.rb
puts "f1.rb"
```

Embed multiple whole files:

```rb
# f1.rb
puts "f1.rb"
```
```go
// f2.go
package main

import "fmt"

func main() {
	fmt.Println("Not embedded")
	// beginembed
	fmt.Println("This is f2.go")
	// endembed
	fmt.Println("Not embedded")
}
```

Embed subsets of a file using `beginembed` and `endembed` magic comments:

```js
// f3.js
console.log("This is f3.js");
```

Embed multiple whole files and multiple subsets of files:

```rb
# f1.rb
puts "f1.rb"
```
```css
/* f4.css */
h1 {
color: blue;
}
```
```html
<!-- f5.html -->
<style>
body {
background-color: #f0f0f0;
}
</style>
```
