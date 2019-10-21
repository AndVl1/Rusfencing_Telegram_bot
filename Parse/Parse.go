package parse

import (
	"golang.org/x/net/html"
	"log"
	"net/http"
)

type Result struct {
	Name, Link string
	Points     string
	Place      string
	Region     string
	Categs     []string
	TeamSquad  map[string]string //team links
}

func ParseLink(link string, isProtocol bool, isTeam bool) []*Result {
	if response, err := http.Get("http://rusfencing.ru" + link); err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Fatal(err)
			} else {
				var items []*Result
				if !isProtocol {
					items = searchFin(doc, "table_block", isProtocol, isTeam)
				} else {
					items = searchFin(doc, "table_block printBody", isProtocol, isTeam)
				}
				return items
			}
		}
	}
	return nil
}

func searchFin(node *html.Node, class string, isProtocol bool, isTeam bool) []*Result {
	if isDiv(node, class) {
		var items []*Result
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Data == "table" {
				items = readTableFin(c.FirstChild.NextSibling.FirstChild, isProtocol, isTeam)
			}
		}
		return items
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if items := searchFin(c, class, isProtocol, isTeam); items != nil {
			return items
		}
	}
	return nil
}
func readTableFin(table *html.Node, isProtocol bool, isTeam bool) []*Result {
	var res []*Result
	for tr := table.NextSibling; tr != nil; tr = tr.NextSibling {
		if tr.Data == "tr" {
			if row := readRowFin(tr, isProtocol, isTeam); row != nil {
				res = append(res, row)
			}
		}

	}
	return res
}

func readRowFin(row *html.Node, isProtocol bool, isTeam bool) *Result {
	for td := row.FirstChild; td != nil; td = td.NextSibling {
		if td.Data == "td" {
			item := readItemFin(td, isProtocol, isTeam)
			if item != nil {
				return item
			}
		}
	}
	return nil
}

func readItemFin(item *html.Node, isProtocol bool, isTeam bool) *Result {
	if a := item.FirstChild; isElem(a, "a") /*|| (isProtocol && isElem(a, "nobr"))*/ {
		cs := getChildren(a)
		if isText(cs[0]) {
			return &Result{
				Name:   cs[0].Data,
				Link:   getAttr(a, "href"),
				Points: getPoints(cs[0]),
				Place:  a.Parent.PrevSibling.PrevSibling.FirstChild.Data,
				Categs: getCat(item),
			}
		} else if isProtocol && (isText(cs[0].FirstChild) || isText(cs[0])) {
			return &Result{
				Name:   cs[0].FirstChild.Data,
				Link:   getAttr(a, "href"),
				Points: "",
				Place:  a.Parent.PrevSibling.PrevSibling.PrevSibling.PrevSibling.FirstChild.Data,
				Categs: nil,
				Region: a.Parent.NextSibling.NextSibling.FirstChild.Data,
			}
		}
	} else if isProtocol {
		if !isTeam && isElem(a, "nobr") {
			return &Result{
				Name:   a.FirstChild.Data,
				Link:   "",
				Points: "",
				Place:  a.Parent.PrevSibling.PrevSibling.PrevSibling.PrevSibling.FirstChild.Data,
				Categs: nil,
				Region: a.Parent.NextSibling.NextSibling.FirstChild.Data,
			}
		} else if isTeam {
			return &Result{
				Name:      a.Parent.NextSibling.NextSibling.FirstChild.Data,
				Link:      "",
				Points:    "",
				Place:     a.Data,
				Region:    "",
				Categs:    nil,
				TeamSquad: getTeam(item),
			}
		}
	}
	return nil
}

func getTeam(item *html.Node) (res map[string]string) {
	res = make(map[string]string)
	defer func(res map[string]string) {
		if x := recover(); x != nil {
			if !isText(item.NextSibling.NextSibling.NextSibling.NextSibling.FirstChild) {
				ch := getChildren(item.NextSibling.NextSibling.NextSibling.NextSibling)
				for _, c := range ch {
					if isElem(c, "a") {
						res[c.FirstChild.Data] = "rusfencing.ru" + getAttr(c, "href")
					}
				}
			} else {
				res[item.NextSibling.NextSibling.NextSibling.NextSibling.FirstChild.Data] = ""
			}
		}
	}(res)
	if !isText(item.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.FirstChild) {
		ch := getChildren(item.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling)
		for _, c := range ch {
			if isElem(c, "a") {
				res[c.FirstChild.Data] = getAttr(c, "href")
			}
		}
	} else {
		res[item.NextSibling.NextSibling.NextSibling.NextSibling.FirstChild.Data] = ""
	}
	return
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
