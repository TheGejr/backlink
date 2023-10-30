package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
	"golang.org/x/net/html"
)

var VERSION = "1.4.0"
var MAX_DEPTH int

func usage() {
	w := os.Stderr
	fmt.Fprintf(w, "Usage: backlink [OPTION...] DOMAIN\n\n")
	fmt.Fprintf(w, "Backlink returns a list of backlinks on a given website - both external and\n")
	fmt.Fprintf(w, "internal resources.\n\n")
	fmt.Fprintf(w, "  Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(w, "\nBacklink v%s was created by Malte Gejr <malte@gejr.dk>\n", VERSION)
}

func main() {
	flag.Usage = usage
	help := flag.BoolP("help", "h", false, "Displays this help message")
	insecure := flag.BoolP("insecure", "k", false, "Allow insecure server connections")
	recursive := flag.BoolP("recursive", "r", false, "Find backlinks recursively on the targeted website")
	max_depth := flag.Int("max-depth", 5, "Max depth for recursive scanning")
	out := flag.StringP("output", "o", "", "Output to a file")

	flag.Parse()
	if *help {
		usage()
		os.Exit(0)
	}

	MAX_DEPTH = *max_depth

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

	current_depth := 1
	res, err := run(domain, current_depth, *insecure, *recursive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encountered: %s\n", err)
	}
	res = removeDuplicateStr(res)

	output(res, out)
}

func run(domain string, current_depth int, insecure bool, recursive bool) ([]string, error) {
	if insecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	resp, err := http.Get(domain)
	if err != nil {
		return []string{}, err
	}

	res := removeDuplicateStr(getLinks(resp.Body, domain))

	// This recursive scanning, could be done alot better
	// but it works and if it ain't broken - don't fix it
	if recursive {
		current_depth = current_depth + 1
		for _, i := range res {
			if current_depth >= MAX_DEPTH {
				break
			}
			if strings.HasPrefix(i, domain) {
				rec_res, err := run(i, current_depth, insecure, recursive)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error encountered: %s\n", err)
				}
				res = append(res, rec_res...)
			}
		}
	}

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
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						if strings.HasPrefix(attr.Val, "#") {
							continue
						}
						if strings.HasPrefix(attr.Val, "/") || strings.HasPrefix(attr.Val, "#") {
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
}
