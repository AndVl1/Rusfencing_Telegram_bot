package parse

import (
	"golang.org/x/net/html"
	"log"
	"net/http"
)

type Compet struct {
	Title, Link string
	Categs      []string
}

type Result struct {
	Place      string
	Name, Link string
}

type Rating struct {
	Place, Link  string
	Name, Points string
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

func ParseResults(link string) []*Result {
	if response, err := http.Get("http://rusfencing.ru" + link); err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		if response.StatusCode == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Fatal(err)
			} else {
				items := searchRes(doc, "table_block printBody")
				return items
			}
		}
	}
	return nil
}

func ParseRatings(link string) []*Rating {
	if response, err := http.Get("http://rusfencing.ru" + link); err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		if response.StatusCode == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Fatal(err)
			} else {
				items := searchRt(doc, "table_block")
				return items
			}
		}
	}
	return nil
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
					Title:  item.Title,
					Link:   item.Link,
					Categs: item.Categs,
				}
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

func search(node *html.Node, class string) []*Compet {
	if isDiv(node, class) {
		var items []*Compet
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Data == "table" {
				items = readTable(c.FirstChild.NextSibling.FirstChild)
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
				Link:   getAttr(a, "href"),
				Title:  cs[0].Data,
				Categs: getCat(item),
			}
		}
	}
	return nil
}
func getPlace(item *html.Node) string {
	res := item.Parent.Parent.PrevSibling.PrevSibling.PrevSibling.PrevSibling.FirstChild.Data
	return res
}

//*****
func searchRt(node *html.Node, class string) []*Rating {
	if isDiv(node, class) {
		var items []*Rating
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Data == "table" {
				items = readTableRt(c.FirstChild.NextSibling.FirstChild)
			}
		}
		return items
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if items := searchRt(c, class); items != nil {
			return items
		}
	}
	return nil
}

func readTableRt(table *html.Node) []*Rating {
	var res []*Rating
	for tr := table.NextSibling; tr != nil; tr = tr.NextSibling {
		if tr.Data == "tr" {
			if row := readRowRt(tr); row != nil {
				res = append(res, row)
			}
		}

	}
	return res
}

func readRowRt(row *html.Node) *Rating {
	for td := row.FirstChild; td != nil; td = td.NextSibling {
		if td.Data == "td" {
			item := readItemRt(td)
			if item != nil {
				return item
			}
		}
	}
	return nil
}

func readItemRt(item *html.Node) *Rating {
	if a := item.FirstChild; isElem(a, "a") {
		cs := getChildren(a)
		if isText(cs[0]) {
			return &Rating{
				Place:  a.Parent.PrevSibling.PrevSibling.FirstChild.Data,
				Link:   getAttr(a, "href"),
				Name:   cs[0].Data,
				Points: getPoints(cs[0]),
			}
		}
	}
	return nil
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

//*****
func searchRes(node *html.Node, class string) []*Result {
	if isDiv(node, class) {
		var items []*Result
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Data == "table" {
				items = readTableRes(c.FirstChild.NextSibling.FirstChild)
			}
		}
		return items
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if items := searchRes(c, class); items != nil {
			return items
		}
	}
	return nil
}

func readRowRes(row *html.Node) *Result {
	for td := row.FirstChild; td != nil; td = td.NextSibling {
		if td.Data == "td" {
			item := readItemRes(td)
			if item != nil {
				return item
			}
		}
	}
	return nil
}

func readTableRes(table *html.Node) []*Result {
	var res []*Result
	for tr := table.NextSibling; tr != nil; tr = tr.NextSibling {
		if tr.Data == "tr" {
			if row := readRowRes(tr); row != nil {
				res = append(res, row)
			}
		}

	}
	return res
}

func readItemRes(item *html.Node) *Result {
	if a := item.FirstChild; isElem(a, "a") {
		cs := getChildren(a)
		if isText(cs[0].FirstChild) {
			return &Result{
				Place: getPlace(cs[0]),
				Link:  getAttr(a, "href"),
				Name:  cs[0].FirstChild.Data,
			}
		}
	}
	return nil
}
