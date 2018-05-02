package main

//go:generate go run main.go version.go commands.go events.go

func main() {
	if err := genVersionFile(); err != nil {
		panic(err)
	}

	if err := generateCommands(); err != nil {
		panic(err)
	}

	if err := generateEvents(); err != nil {
		panic(err)
	}
}
