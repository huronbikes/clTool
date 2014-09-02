package main

import (
	"fmt"
	"net/http"
	"html"
	"encoding/json"
	"github.com/clTool/clSearch"
	"strings"
)


var SRCH *clSearch.Search

func searchhandler(w http.ResponseWriter, r *http.Request) {
	paths:=strings.Split(r.URL.Path, "/")
	if len(paths) > 0 {
		for {
			if len(paths) < 1 || paths[0] != "" {
				break
			}
			paths = paths[1:]
		}
	}

	if len(paths) >= 3 {
		SRCH = new(clSearch.Search)
		sect := paths[1]
		query := paths[2]
		SRCH.Init(sect, query)
		SRCH.SearchCL()
		http.Redirect(w,r,"/",302)
	}
}

func nexthandler(w http.ResponseWriter, r *http.Request) {
	if SRCH != nil {
		SRCH.NextPage()
	}
	http.Redirect(w,r,"/",302)
}

func currentpage(w http.ResponseWriter, r *http.Request) {
	enc:=json.NewEncoder(w)
	enc.Encode(SRCH.GetCurrentPage())
}

func handler(w http.ResponseWriter, r *http.Request) {
	var currentSearch string
	if SRCH != nil {
		currentSearch=SRCH.Query
	}
	fmt.Fprintf(w,"<html><head><title>CL General Search</title></head>")
	fmt.Fprintf(w, "<body>")
	code := "var sect=document.getElementById('sect').value;var query=document.getElementById('query').value;window.location='/search/'+sect+'/'+query; return false;"
	fmt.Fprintf(w, "<form onsubmit=\"%s\">", code)
	fmt.Fprintf(w, "<label for=\"sect\">Section</label><input type=\"text\" id=\"sect\" name=\"sect\" disabled=\"yes\" value=\"mcy\"></input>")
	fmt.Fprintf(w, "<label for=\"sect\">Search</label><input type=\"text\" id=\"query\" name=\"query\" value=\"%s\"></input>", html.EscapeString(currentSearch))
	fmt.Fprintf(w, "<input type=\"submit\" onclick=\"%s\" name=\"Search\"></input>", code)
	fmt.Fprintf(w, "</form>")
	if (SRCH == nil) {
		fmt.Fprintf(w,"<h4>No search in progress</h4>")
	} else {
		fmt.Fprintf(w, "<ul>")
		fmt.Fprintf(w, "<li style=\"width:100%%;\">")
		fmt.Fprintf(w,"<span style=\"width:25%%; display: block; float: left;\">Source Craigslist Site</span><span>Post</span></li>")

		for _, i := range SRCH.GetCurrentPage().Links {
			fmt.Fprintf(w, "<li style=\"width:100%%;\">")
			fmt.Fprintf(w,"<span style=\"width:25%%; display: block; float: left;\">")
			fmt.Fprintf(w, "%s</span><a href=\"%s\">%s (%s)</a></li>", i.Url.Host, i.GetLinkUrl(), i.PostTitle, i.Price)
		}
		fmt.Fprintf(w, "</ul>")
		fmt.Fprintf(w,"<a href=\"/next/\">Next</a>")
	}

	fmt.Fprintf(w, "</body>")
	fmt.Fprintf(w,"</html>")
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/next/",  nexthandler)
	http.HandleFunc("/search/", searchhandler)
	http.HandleFunc("/current/", currentpage)
	http.ListenAndServe(":6080", nil)
}
