package clSearch

import (
	"fmt"
	"strings"
	"net/http"
	"code.google.com/p/go.net/html"
	"sync"
	"net/url"
)


const DefaultPageSize int = 25

type Page struct {
	Links []Link
}

type Search struct {
	CLUrls      chan string
	CLResults   chan Link
	Section     string
	Query       string
	PageSize int
	currentPage *Page
	resultMutex sync.Mutex
	workers workerPool
	searchState searchState
}


func (srch *Search)transitionSearchState(state int){
	srch.searchState.setSearchState(state)
	switch state {
	case SearchCompleted:
		close(srch.CLResults)
	}
}

func (srch *Search) FillPage () *Page {
	srch.resultMutex.Lock()
	defer srch.resultMutex.Unlock()
	page := new(Page);
	page.Links = make([]Link, 0, srch.PageSize)
	for i:=0; i < srch.PageSize; i++ {
		result, ok := <- srch.CLResults
		if ok {
			page.Links = append(page.Links, result)
		} else {
			break
		}
	}
	return page
}

func (srch *Search) NextPage () {
	srch.currentPage = srch.FillPage()
}

func (srch *Search) GetCurrentPage () *Page {
	if srch.currentPage == nil {
		srch.NextPage();
	}
	return srch.currentPage
}

func (srch *Search) LinkFromElement (node *html.Node, baseUrl *url.URL) *Link {
	attr := GetAttr(node, "data-pid")
	if attr != nil {
		link:=new(Link)
		link.Section = srch.Section
		link.PostId=attr.Val
		findfn:=func(nd *html.Node) bool {
			return nd.Data=="span" && HasAttr(nd, "class", "price")
		}
		if node.FirstChild !=nil {
			price := Find(node.FirstChild, findfn)
			if price != nil && price.FirstChild != nil {
				link.Price = price.FirstChild.Data
			}
			title := Find(node.FirstChild, func(nd *html.Node) bool {
					return nd.Data == "a" && HasAttr(nd, "class", "hdrlnk")
				})

			if title != nil && title.FirstChild != nil {
				link.OriginalHref = GetAttr(title, "href").Val
				hrefUrl,_ := url.Parse(link.OriginalHref)
				if !hrefUrl.IsAbs() {
					hrefUrl.Host=baseUrl.Host
					hrefUrl.Scheme=baseUrl.Scheme
				}
				link.Url=hrefUrl
				link.PostTitle = title.FirstChild.Data
			}
		}
		return link
	}
	return nil
}

/*
	Initialization method for Seach.
 */

func (srch *Search) Init(Sect string, Query string) {
	srch.CLUrls = make(chan string, 5)
	srch.CLResults = make(chan Link, 5)
	srch.PageSize = DefaultPageSize
	srch.Section = Sect
	srch.Query = Query
	srch.workers.init()
	srch.transitionSearchState(SearchInitialized)
}

func Traverse(doc *html.Node, tvsfn func(*html.Node)) {
	tvsfn(doc)
	if doc.NextSibling != nil {
		Traverse(doc.NextSibling, tvsfn)
	}
	if doc.FirstChild != nil {
		Traverse(doc.FirstChild, tvsfn)
	}
}


func (srch *Search)Dispose() {
	srch.workers.stop()
	for _ = range srch.CLResults {
	}
}

func (srch *Search) _getLinks(doc *html.Node) {
	tvsfn := func(doc *html.Node) {
		if doc.Type == html.ElementNode && doc.Data == "a" {
			for _, attr := range doc.Attr {
				if attr.Key == "href" &&
						(strings.HasPrefix(strings.ToLower(attr.Val), "https://") ||
								strings.HasPrefix(strings.ToLower(attr.Val), "http://")) {
					srch.CLUrls <- attr.Val;
					break;
				}
			}
		}
	}
	Traverse(doc, tvsfn)
}

func (srch *Search) getLinks(doc *html.Node) {
	srch._getLinks(doc)
	close(srch.CLUrls)
}

func (srch *Search) SearchCL() {
	srch.transitionSearchState(SearchInProgress)
	srch.workers.addWorker(func() {
		srch._searchCl()
	})
}

func (srch *Search) _searchCl() {
	resp, err := http.Get("http://geo.craigslist.org/iso/us/")
	defer srch.workers.workerCompleted()
	if err == nil {
		doc, _ := html.Parse(resp.Body);
		go srch.getLinks(doc)
		for {
			clBaseUrl, ok := <-srch.CLUrls
			if !ok {
				return
			}
			srch.workers.addWorker(func() {
				srch.GetResults(clBaseUrl)
			})
		}
	} else {
		close(srch.CLResults)
		close(srch.CLUrls)
		fmt.Println(err)
	}
}

func (srch *Search) maybeCompleted(){
	srch.workers.workerCompleted()
	if srch.workers.workerCount() == 0 {
		srch.transitionSearchState(SearchCompleted)
	}
}

func (srch *Search) GetResults(clBaseUrl string) {
	defer srch.maybeCompleted()
	clurl,_ := url.Parse(clBaseUrl)
	clurl.Path="/search/" + srch.Section
	query := clurl.Query()
	query.Set("query",srch.Query)
	clurl.RawQuery=query.Encode()


	resp, err := http.Get(clurl.String());
	if err != nil {
		fmt.Print(err)
		return
	}
	doc, _ := html.Parse(resp.Body);
	findfn := func(doc *html.Node) bool {
		return doc.Type == html.ElementNode && doc.Data == "div" && HasAttr(doc, "class", "content")
	}
	if doc == nil {
		return
	}
	content := Find(doc, findfn);
	if content != nil {
		link := content.FirstChild
		for {
			Link := srch.LinkFromElement(link, clurl)
			if Link != nil {
				srch.CLResults <- *Link
			}
			if srch.workers.stopped() {
				return
			}
			link = link.NextSibling
			if link == nil || (link.Type == html.ElementNode && link.Data == "h4" && HasAttr(link, "class", "ban nearby")) {
				break
			}
		}
	}
}
