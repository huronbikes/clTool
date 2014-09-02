package clSearch

import (
	"fmt"
	"strings"
	"net/http"
	"code.google.com/p/go.net/html"
	"sync"
)

const PageSize int = 25

type Page struct {
	Links []Link
}

type Search struct {
CLUrls    chan string
CLResults chan Link
Section   string
Query     string
currentPage *Page
resultMutex sync.Mutex
workerMutex sync.Mutex
workerCount int
bool
}

type Link struct {
	BaseUrl string
	Section string
	PostId string
	PostTitle string
	Price string
}

func (srch *Search) FillPage () *Page {
	srch.resultMutex.Lock()
	defer srch.resultMutex.Unlock()
	page := new(Page);
	page.Links = make([]Link, PageSize, PageSize)
	fmt.Println("Filling page ",srch.WorkerCount())
	for i:=0; i < PageSize; i++ {
		if srch.WorkerCount() > 0 {
			page.Links[i] = <- srch.CLResults
		}
	}
	return page
}

func (srch *Search) AddWorker() {
	srch.workerMutex.Lock()
	defer srch.workerMutex.Unlock()
	fmt.Println("Starting worker... ", srch.workerCount)
	srch.workerCount++
}

func (srch *Search) WorkerCount() int{
	srch.workerMutex.Lock()
	defer srch.workerMutex.Unlock()
	return srch.workerCount
}

func (srch *Search) WorkerCompleted() {
	srch.workerMutex.Lock()
	defer srch.workerMutex.Unlock()
	fmt.Println("Ending worker... ",srch.workerCount)
	srch.workerCount--
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


func (lnk *Link) GetLinkUrl () string {
	clBaseUrl := strings.TrimSuffix(lnk.BaseUrl, "/")
	url := fmt.Sprintf("%s/%s/%s.html", clBaseUrl, lnk.Section, lnk.PostId)
	return url
}

func (lnk *Link) PostString () string {
	return fmt.Sprintf("%s %s %s", lnk.GetLinkUrl(), lnk.PostTitle, lnk.Price)
}

func (srch *Search) LinkFromElement (node *html.Node, baseUrl string) *Link {
	attr := GetAttr(node, "data-pid")
	if attr != nil {
		link:=new(Link)
		link.BaseUrl=baseUrl
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
				link.PostTitle = title.FirstChild.Data
			}
		}
		return link
	}
	return nil
}

func (srch *Search) Init(Sect string, Query string) {
	srch.CLUrls = make(chan string, 5)
	srch.CLResults = make(chan Link, 5)
	srch.Section = Sect
	srch.Query = Query
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

/*
   find the first node such that findfn returns a value
 */
func Find(doc *html.Node, findfn func(*html.Node) bool) *html.Node {
	result := findfn(doc)
	var found *html.Node
	if result {
		return doc
	}
	if doc.NextSibling != nil && found == nil {
		found = Find(doc.NextSibling, findfn)
	}
	if doc.FirstChild != nil && found == nil {
		found = Find(doc.FirstChild, findfn)
	}
	return found
}

/*
  find the first attribute with named by attrKey
 */
func GetAttr(doc *html.Node, attrKey string) *html.Attribute {
	attrs := doc.Attr
	for _, i := range attrs {
		if i.Key == attrKey {
			return &i
		}
	}
	return nil
}

/*
  Returns true if Node doc has an attribute with key attr.Key and value attr.Val
 */

func HasAttr(doc *html.Node, attrKey string, attrVal string) bool {
	testAttr := GetAttr(doc, attrKey)
	if testAttr != nil {
		return testAttr.Val == attrVal
	}
	return false
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
	defer srch.WorkerCompleted()
	close(srch.CLUrls)
}

func (srch *Search) SearchCL() {
	srch.AddWorker()
	go srch._searchCl()
}

func (srch *Search) _searchCl() {
	resp, _ := http.Get("http://geo.craigslist.org/iso/us/")
	doc, _ := html.Parse(resp.Body);
	go srch.getLinks(doc)
	for {
		clBaseUrl, ok := <-srch.CLUrls
		if !ok {
			return
		}
		srch.AddWorker()
		go srch.GetResults(clBaseUrl)
	}
}

func (srch *Search) GetResults(clBaseUrl string) {
	defer srch.WorkerCompleted()
	clBaseUrl = strings.TrimSuffix(clBaseUrl, "/")
	url := fmt.Sprintf("%s/search/%s?query=%s", clBaseUrl, srch.Section, srch.Query)
	resp, err := http.Get(url);
	if err != nil {
		fmt.Print(err)
		return

	}
	doc, _ := html.Parse(resp.Body);
	findfn := func(doc *html.Node) bool {
		return doc.Type == html.ElementNode && doc.Data == "div" && HasAttr(doc, "class", "content")
	}

	content := Find(doc, findfn);
	if content != nil {
		link := content.FirstChild
		for {
			Link := srch.LinkFromElement(link, clBaseUrl)
			if Link != nil {
				srch.CLResults <- *Link
			}
			link = link.NextSibling
			if link == nil || (link.Type == html.ElementNode && link.Data == "h4" && HasAttr(link, "class", "ban nearby")) {
				break
			}
		}
	}
}
