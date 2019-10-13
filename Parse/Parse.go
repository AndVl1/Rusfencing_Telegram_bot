package parse

import (
	"golang.org/x/net/html"
	"log"
	"net/http"
)

type Compet struct {
	Title, Link string
}

type fields struct {
	vozr 	[]int
	weapon 	[]int
	sex 	[]int
	vid 	[]int
}

func ParseCompetitions() []*Compet {
	link := "rusfencing.ru/result.php"
	if response, err := http.Get("http://" + link); err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Fatal(err)
			} else {
				items := search(doc, "table_block")
				return items
			}
		}
	}
	return nil
}

func ParseResults() {

}

func readTable(table *html.Node) []*Compet {
	var res []*Compet
	for tr := table.NextSibling; tr != nil; tr = tr.NextSibling {
		if tr.Data == "tr" {
			if row := readRow(tr); row != nil {
				res = append(res, row)
			}
		}

	}
	return res
}

func readRow(row *html.Node) *Compet {
	for td := row.FirstChild; td != nil; td = td.NextSibling {
		if td.Data == "td" {
			item := readItem(td)
			if item != nil {
				return &Compet{
					Title: item.Title,
					Link:  item.Link,
				}
			}
		}
	}
	return nil
}

func search(node *html.Node, class string) []*Compet {
	if isDiv(node, class) {
		log.Printf("==== %s ====", class)
		var items []*Compet
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Data == "table" {
				items = readTable(c.FirstChild.NextSibling.FirstChild)
				//tr := c.FirstChild.NextSibling.FirstChild.NextSibling.NextSibling
				//log.Println(tr.Data)
				//item := readItem(c)
				//if item != nil {
				//	log.Println("appending, len = ", len(items))
				//	items = append(items, item)
				//}
			}
		}
		return items
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if items := search(c, class); items != nil {
			return items
		}
	}
	return nil
}

func getChildren(node *html.Node) []*html.Node {
	var children []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}
	return children
}

func getAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func isText(node *html.Node) bool {
	return node != nil && node.Type == html.TextNode
}

func isTable(node *html.Node, tag string) bool {
	return node != nil && isElem(node, "table")
}

func isElem(node *html.Node, tag string) bool {
	return node != nil && node.Type == html.ElementNode && node.Data == tag
}

func isDiv(node *html.Node, class string) bool {
	return isElem(node, "div") && getAttr(node, "class") == class
}

func readItem(item *html.Node) *Compet {
	if a := item.FirstChild; isElem(a, "a") {
		cs := getChildren(a)
		if isText(cs[0]) {
			return &Compet{
				Link: getAttr(a, "href"),
				Title: cs[0].Data,
			}
		}
	}
	return nil
}

