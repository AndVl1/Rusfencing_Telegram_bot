package parse

import (
	"golang.org/x/net/html"
	"log"
	"net/http"
)

type ResultFin struct {
	Name, Link string
	Points     string
	Place      string
	Categs     []string
}

func ParseLink(link string, isProtocol bool) []*ResultFin {
	if response, err := http.Get("http://rusfencing.ru" + link); err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Fatal(err)
			} else {
				var items []*ResultFin
				if !isProtocol {
					items = searchFin(doc, "table_block", isProtocol)
				} else {
					items = searchFin(doc, "table_block printBody", isProtocol)
				}
				return items
			}
		}
	}
	return nil
}

func searchFin(node *html.Node, class string, isProtocol bool) []*ResultFin {
	if isDiv(node, class) {
		var items []*ResultFin
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Data == "table" {
				items = readTableFin(c.FirstChild.NextSibling.FirstChild, isProtocol)
			}
		}
		return items
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if items := searchFin(c, class, isProtocol); items != nil {
			return items
		}
	}
	return nil
}
func readTableFin(table *html.Node, isProtocol bool) []*ResultFin {
	var res []*ResultFin
	for tr := table.NextSibling; tr != nil; tr = tr.NextSibling {
		if tr.Data == "tr" {
			if row := readRowFin(tr, isProtocol); row != nil {
				res = append(res, row)
			}
		}

	}
	return res
}

func readRowFin(row *html.Node, isProtocol bool) *ResultFin {
	for td := row.FirstChild; td != nil; td = td.NextSibling {
		if td.Data == "td" {
			item := readItemFin(td, isProtocol)
			if item != nil {
				return item
			}
		}
	}
	return nil
}

func readItemFin(item *html.Node, isProtocol bool) *ResultFin {
	if a := item.FirstChild; isElem(a, "a") {
		cs := getChildren(a)
		if isText(cs[0]) {
			return &ResultFin{
				Name:   cs[0].Data,
				Link:   getAttr(a, "href"),
				Points: getPoints(cs[0]),
				Place:  a.Parent.PrevSibling.PrevSibling.FirstChild.Data,
				Categs: getCat(item),
			}
		} else if isProtocol && isText(cs[0].FirstChild) {
			return &ResultFin{
				Name:   cs[0].FirstChild.Data,
				Link:   getAttr(a, "href"),
				Points: "",
				Place:  a.Parent.PrevSibling.PrevSibling.PrevSibling.PrevSibling.FirstChild.Data,
				Categs: nil,
			}
		}
	}
	return nil
}

func getCat(item *html.Node) []string {
	var result []string
	for data := item.NextSibling; data != nil; data = data.NextSibling {
		if cat := data.FirstChild; cat != nil {
			if cat.Data == "Личные" || cat.Data == "Командные" || cat.Data == "М" || cat.Data == "Ж" || cat.Data == "сабля" || cat.Data == "шпага" || cat.Data == "рапира" {
				result = append(result, cat.Data)
			}
		}
	}
	return result
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

func isElem(node *html.Node, tag string) bool {
	return node != nil && node.Type == html.ElementNode && node.Data == tag
}

func isDiv(node *html.Node, class string) bool {
	return isElem(node, "div") && getAttr(node, "class") == class
}

func getPoints(item *html.Node) string {
	res := ""
	for cur := item.Parent.Parent; cur != nil; cur = cur.NextSibling {
		if cur.FirstChild != nil && isText(cur.FirstChild) {
			res = cur.FirstChild.Data
		}
	}
	return res
}
