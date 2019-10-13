package main

import (
	"fmt"
	parse "rfgTgBot/Parse"
)

func main() {
	//ps := parse.ParseCompetitions()
	//for _, v := range ps {
	//	fmt.Println(v.Title, v.Link, v.Categs)
	//}
	ps2 := parse.ParseResults("/protocol.php?ID=2029095")
	for _, v := range ps2 {
		fmt.Println(v.Name, v.Link)
	}
}
