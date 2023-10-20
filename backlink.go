package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
	"golang.org/x/net/html"
)

var VERSION = "1.0.0"

func usage() {
	w := os.Stderr
	fmt.Fprintf(w, "Usage: backlink [OPTION...] DOMAIN\n\n")
	fmt.Fprintf(w, "Backlink returns a list of backlinks on a given website - both external and\n")
	fmt.Fprintf(w, "internal resources.\n\n")
	fmt.Fprintf(w, "  Options:\n")
	// fmt.Fprintf(w, "    -h, --help\tDisplays this help message\n")
	flag.PrintDefaults()
	fmt.Fprintf(w, "\nBacklink v%s was created by Malte Gejr <malte@gejr.dk>\n", VERSION)
}

func main() {
	flag.Usage = usage
	help := flag.BoolP("help", "h", false, "Displays this help message")
	out := flag.StringP("output", "o", "", "Output to a file")

	flag.Parse()
	if *help {
		usage()
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		fmt.Printf("domain is required\n\n")
		usage()
		os.Exit(1)
	}
	domain := os.Args[1]
	_, err := url.ParseRequestURI(domain)
	if err != nil {
		fmt.Printf("domain must be a valid domain, starting with http:// or https://\n\n")
		usage()
		os.Exit(1)
	}

	res, err := run(domain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encountered: %s", err)
	}

	output(res, out)
}

func run(domain string) ([]string, error) {
	resp, err := http.Get(domain)
	if err != nil {
		return []string{}, err
	}

	res := removeDuplicateStr(getLinks(resp.Body, domain))

	return res, nil
}

func getLinks(body io.Reader, domain string) []string {
	var links []string
	z := html.NewTokenizer(body)
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if "a" == token.Data {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						if strings.HasPrefix(attr.Val, "/") {
							attr.Val = fmt.Sprintf("%s%s", domain, attr.Val)
						}
						links = append(links, attr.Val)
					}

				}
			}

		}
	}
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func output(output []string, out *string) {
	if *out == "" {
		for _, v := range output {
			fmt.Fprintln(os.Stdout, v)
		}
	} else {
		w, err := os.Create(*out)
		if err != nil {
			panic(err)
		}
		for _, v := range output {
			fmt.Fprintln(w, v)
		}
	}
	// TODO: Add outputting to a file
}
