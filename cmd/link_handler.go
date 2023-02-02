package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func (a *app) handleLink(link string) (*Bill, error) {
	content, err := getHtml(link)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	billContent := findBill(doc)

	bill, err := parseBil(billContent)
	if err != nil {
		return nil, err
	}

	return bill, nil
}

func getHtml(link string) (string, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()
	body, err := ioutil.ReadAll(res.Body)
	return string(body), err
}

func parseBil(billContent string) (*Bill, error) {
	lines := strings.Split(billContent, "\n")
	var items []Item
	itemsIndex := -1
	var totalAmount int64
	var boughtAt time.Time
	var err error

	for i, line := range lines {
		if strings.HasPrefix(line, "Назив") {
			itemsIndex = i + 1
		}
		if i == itemsIndex {
			additionalTitleLine := 0
			if !strings.HasPrefix(lines[i+1], " ") {
				additionalTitleLine = 1
			}
			item, err := parseItem(lines[i+1+additionalTitleLine])
			if err != nil {
				return nil, err
			}
			item.Name = strings.TrimSpace(line)
			items = append(items, *item)
			if strings.HasPrefix(lines[i+2+additionalTitleLine], "--------") {
				itemsIndex = -1
			} else {
				itemsIndex = itemsIndex + 2 + additionalTitleLine
			}
		}

		if strings.HasPrefix(line, "Укупан износ:") {
			valueStr := strings.ReplaceAll(line, "Укупан износ:", "")
			valueStr = strings.TrimSpace(valueStr)
			valueStr = strings.ReplaceAll(valueStr, ".", "")
			valueStr = strings.ReplaceAll(valueStr, ",", "")
			totalAmount, err = strToInt(valueStr)
			if err != nil {
				return nil, err
			}
		}
		// ПФР време:
		if strings.HasPrefix(line, "ПФР време:") {
			valueStr := strings.ReplaceAll(line, "ПФР време:", "")
			valueStr = strings.TrimSpace(valueStr)
			var err error
			boughtAt, err = time.Parse("02.01.2006. 15:04:05", valueStr)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	return &Bill{
		TotalAmount: totalAmount,
		BoughtAt:    boughtAt,
		Description: "Супермаркет",
		Category:    "Продукты",
		Items:       items,
	}, nil
}

func parseItem(article string) (*Item, error) {
	article = strings.ReplaceAll(strings.TrimSpace(article), ".", "")
	item := Item{}
	start := 0
	var err error
	prevIsSpace := false
	k := 0
	for i := 0; i < len(article); i++ {
		c := article[i]
		isSpace := unicode.IsSpace(rune(c))
		if isSpace && !prevIsSpace {
			if k == 0 {
				item.Price, err = strToInt(strings.ReplaceAll(article[start:i], ",", ""))
				if err != nil {
					return nil, err
				}
				k = 1
			} else if k == 1 {
				item.Count, err = strToFloat(strings.ReplaceAll(article[start:i], ",", "."))
				if err != nil {
					return nil, err
				}
			}
		}
		if !isSpace && prevIsSpace {
			start = i
		}
		prevIsSpace = isSpace
	}
	item.Sum, err = strToInt(strings.ReplaceAll(article[start:], ",", ""))
	if err != nil {
		return nil, nil
	}
	return &item, nil
}

func strToInt(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

func strToFloat(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

func findBill(doc *html.Node) string {
	var traverse func(n *html.Node) string
	traverse = func(n *html.Node) string {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if strings.HasPrefix(c.Data, "============ ФИСКАЛНИ РАЧУН") {
				return c.Data
			}
			res := traverse(c)
			if res != "" {
				return res
			}
		}
		return ""
	}
	return traverse(doc)
}
