package main

import (
	"fmt"
	"net/http"
	"github.com/clSearch"
	"strings"
)


var SRCH *clSearch.Search

func handler(w http.ResponseWriter, r *http.Request) {
	paths:=strings.Split(r.URL.Path, "/")
	if len(paths) > 0 {
		for {
			if len(paths) < 1 || paths[0] != "" {
				break
			}
			paths = paths[1:]
		}
	}
	fmt.Println(paths)
	if len(paths) > 0 {
		switch strings.ToLower(paths[0]) {
			case "search":
				if len(paths) >= 3 {
					SRCH = new(clSearch.Search)
					sect := paths[1]
					query := paths[2]
					SRCH.Init(sect, query)
					go SRCH.SearchCL()
					fmt.Println(paths)
					http.Redirect(w,r,"/",302)
				}
			case "next":
				if SRCH != nil {
					SRCH.NextPage()
				}
				http.Redirect(w,r,"/",302)
		}
	}


	fmt.Fprintf(w,"<html><head><title>CL General Search</title></head>")
	fmt.Fprintf(w, "<body>")

	if (SRCH == nil) {
		fmt.Fprintf(w,"<h4>No search in progress</h4>")
	} else {
		fmt.Fprintf(w, "<ul>")
		for _, i := range SRCH.GetCurrentPage().Links {
			fmt.Fprintf(w, "<li><span>%s</span><a href=\"%s\">%s (%s)</a></li>", i.BaseUrl, i.GetLinkUrl(), i.PostTitle, i.Price)
		}
		fmt.Fprintf(w, "</ul>")
		fmt.Fprintf(w,"<a href=\"/next\">Next</a>")
	}

	fmt.Fprintf(w, "</body>")
	fmt.Fprintf(w,"</html>")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":6080", nil)
}

/*
func main() {
	var srch = new(clSearch)
	srch.Init("mcy", "cafe")


	for i := 0; i < 25; i++ {
		i := <- srch.CLResults
		fmt.Println(i.PostString())
	}
	//srch.waitgroup.Wait()
}

*/
