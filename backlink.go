package main

// Issue 01:
// The $domain argument must be passed as the very first argument, if options is present before the domain
// `backlink` will panic because it's hardcoded to use `os.Args[1]` as domain.

// Issue 02:
// Sometimes without the recursive flag, it does a recursive scan with DEPTH=1

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	flag "github.com/spf13/pflag"
	"golang.org/x/net/html"
)

var VERSION = "1.4.1"

type Backlink struct {
	CurrentDepth    int
	LocalDomain     string
	Uri             url.URL
	LocalResources  map[string]bool
	ExternResources map[string]bool
	Options         Options
}

type Options struct {
	MaxDepth  int
	Insecure  bool
	Recursive bool
}

var WHITELIST_SCHEME = []string{"http", "https", ""} // empty string to capture backlinks like "/news"

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

	if len(os.Args) < 2 {
		fmt.Printf("domain is required\n\n")
		usage()
		os.Exit(1)
	}
	domain := os.Args[1]
	url_obj, err := url.ParseRequestURI(domain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "domain must be a valid domain, starting with http:// or https://\n\n")
		usage()
		os.Exit(1)
	}

	backlink := Backlink{
		CurrentDepth:    1,
		Uri:             *url_obj,
		LocalDomain:     url_obj.Scheme + "://" + url_obj.Host,
		LocalResources:  make(map[string]bool),
		ExternResources: make(map[string]bool),
		Options: Options{
			MaxDepth:  *max_depth,
			Insecure:  *insecure,
			Recursive: *recursive,
		},
	}
	backlink.LocalResources[url_obj.Path] = false

	err = backlink.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encountered: %s\n", err)
	}

	backlink.Output(out)
}

func (b Backlink) Run() error {
	if b.Options.Insecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	for resource, scanned := range b.LocalResources {
		if !scanned {
			b.LocalResources[resource] = true // Resource now scanned
			resp, err := http.Get(b.LocalDomain + resource)
			if err != nil {
				return err
			}

			backlinks := removeDuplicateStr(getLinks(resp.Body))
			for _, backlink := range backlinks {
				url_obj, err := url.ParseRequestURI(backlink)
				if err != nil {
					return err
				}
				if !slices.Contains(WHITELIST_SCHEME, url_obj.Scheme) {
					continue
				}

				if url_obj.Host == b.Uri.Host || url_obj.Host == "" || url_obj.Host == "www."+b.Uri.Host {
					_, exist := b.LocalResources[url_obj.Path]
					if exist {
						continue
					}
					b.LocalResources[url_obj.Path] = false
				} else {
					b.ExternResources[backlink] = true
				}
			}
		}
	}
	if b.Options.Recursive {
		for {
			b.CurrentDepth++
			if b.CurrentDepth >= b.Options.MaxDepth {
				break
			}

			err := b.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encountered: %s\n", err)
			}
		}
	}

	return nil
}

func getLinks(body io.Reader) []string {
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
						if strings.HasPrefix(attr.Val, "#") { // Drop # anchors
							continue
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

func (b Backlink) Output(out *string) {
	if *out == "" {
		for resource, _ := range b.LocalResources {
			fmt.Fprintln(os.Stdout, b.LocalDomain+resource)
		}
	} else {
		w, err := os.Create(*out)
		if err != nil {
			panic(err)
		}
		for resource, _ := range b.LocalResources {
			fmt.Fprintln(w, b.LocalDomain+resource)
		}
	}
}
