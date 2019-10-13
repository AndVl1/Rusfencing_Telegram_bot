package main

import (
	"fmt"
	parse "rfgTgBot/Parse"
)

func main() {
	ps := parse.ParseCompetitions()
	for _, v := range ps {
		fmt.Println(v.Title, v.Link)
	}
}
