package main

import (
	m "davidhancock.com/marmot/marmot"
)

func main() {

	m.ParseArguments() // guarantees that arguments are acceptable
	
	m.GoForIt()

}
