package main

import "fmt"

func main() {
	fmt.Println("Not embedded")
	// beginembed
	fmt.Println("This is f2.go")
	// endembed
	fmt.Println("Not embedded")
}
