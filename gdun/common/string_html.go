package common

import (
	"container/list"
	"golang.org/x/net/html"
	"strings"
)

//----------------------------------------------------------------------------

func HTML2PlainText(s string, defaultDomain string) (string, error) {
	q := strings.Replace(s, "\r", "", -1)
	q = strings.Replace(q, "\n", "", -1)
	q = strings.Replace(q, "\t", " ", -1)
	q = strings.Replace(q, "+", "%2B", -1)

	node, err := html.Parse(strings.NewReader(q))
	if err != nil {
		return "", err
	}

	tmp := list.New()
	final := list.New()
	parse(node, tmp, final, defaultDomain)
	saveTmpList(tmp, final)

	r := ``
	first := true
	for e := final.Front(); e != nil; e = e.Next() {
		line, okay := e.Value.(string)
		if !okay {
			continue
		}

		if len(line) == 0 {
			continue
		}

		if first {
			first = false
		} else {
			r += "\\n"
		}
		r += line
	}

	return r, nil
}

func parse(node *html.Node, tmp *list.List, final *list.List, defaultDomain string) {

	// Parse this node.
	switch node.Type {

	case html.TextNode:
		s := strings.TrimSpace(node.Data)
		if len(s) > 0 {
			tmp.PushBack(s)
		}

	case html.ElementNode:
		name := strings.ToLower(node.Data)
		switch name {
		case "img":
			// Combine temporary lines and save the result to the list.
			saveTmpList(tmp, final)

			// Get its source.
			for i := 0; i < len(node.Attr); i++ {
				if node.Attr[i].Key == "src" {
					url := node.Attr[i].Val
					n := len(url)
					if n > 0 {
						if url[0] == '/' {
							if (n > 1) && (url[1] == '/') {
								url = "https:" + url
							} else {
								url = "https://" + defaultDomain + url
							}
						}
						final.PushBack(url)
					}
					break
				}
			}

		case "br":
			// Combine temporary lines and save the result to the list.
			saveTmpList(tmp, final)

		case "p", "table", "thead", "tr":
			// Combine temporary lines and save the result to the list.
			saveTmpList(tmp, final)

			// Parse its children.
			parseChildren(node, tmp, final, defaultDomain)
			saveTmpList(tmp, final)

		default:
			parseChildren(node, tmp, final, defaultDomain)
		}

	case html.DocumentNode:
		parseChildren(node, tmp, final, defaultDomain)
	}
}

func saveTmpList(tmp *list.List, final *list.List) {
	r := ""
	for e := tmp.Front(); e != nil; e = e.Next() {
		s, okay := e.Value.(string)
		if !okay {
			continue
		}

		if len(s) == 0 {
			continue
		}

		if len(r) == 0 {
			r = s
		} else {
			r += " " + s
		}
	}

	tmp.Init()

	if len(r) > 0 {
		final.PushBack(r)
	}
}

func parseChildren(node *html.Node, tmp *list.List, final *list.List, defaultDomain string) {
	for p := node.FirstChild; p != nil; p = p.NextSibling {
		parse(p, tmp, final, defaultDomain)
	}
}

func parseAppendLine(s string, ls *list.List) {
	t := strings.TrimSpace(s)
	if len(t) > 0 {
		ls.PushBack(t)
	}
}

func parseConcatLine(s string, ls *list.List) {
	t := strings.TrimSpace(s)
	if len(t) == 0 {
		return
	}

	if ls.Len() == 0 {
		ls.PushBack(t)
		return
	}

	// Get the last line.
	e := ls.Back()
	line, okay := e.Value.(string)
	if !okay {
		return
	}
	ls.Remove(e)

	line += " " + t
	ls.PushBack(line)
}

//----------------------------------------------------------------------------
