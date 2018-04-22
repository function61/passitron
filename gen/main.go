package main

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
