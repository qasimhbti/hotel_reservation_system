package main

func main() {
	a := App{}
	a.Initialize("root", "root", "restaurant_api")
	a.Run(":8080")
}
