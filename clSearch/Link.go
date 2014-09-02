package clSearch

import (
"fmt"
"net/url"
)

type Link struct {
	Url *url.URL
	OriginalHref string
	Section string
	PostId string
	PostTitle string
	Price string
}

func (lnk *Link) GetLinkUrl () string {
	return lnk.Url.String()
}

func (lnk *Link) PostString () string {
	return fmt.Sprintf("%s %s %s", lnk.GetLinkUrl(), lnk.PostTitle, lnk.Price)
}
