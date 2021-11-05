package blogcluster

import (
	"io"
	"log"
	nurl "net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
)

// https://www.cnblogs.com/taoshihan/p/14389040.html
// https://zhuanlan.zhihu.com/p/33122909?from_voters_page=true
// https://cloud.tencent.com/developer/article/1066274
// https://zhuanlan.zhihu.com/p/80213099
// https://blog.csdn.net/RobertChenGuangzhi/article/details/108068695
// https://www.jianshu.com/p/7ca8ce9746fb
// https://stackoverflow.com/questions/16895294/how-to-set-timeout-for-http-get-requests-in-golang
// https://www.golangprograms.com/regular-expression-to-extract-date-yyyy-mm-dd-from-string.html
// golangbyexample.com/parse-time-in-golang
// https://www.geeksforgeeks.org/how-to-compare-times-in-golang/
// https://www.cnblogs.com/roastpiglet/p/12123524.html

// 解析器接口。在解析模块当中进行具体实现。
type Parser interface {

	// 解析源码方法。输入网页地址和html源码，输出网页纯文本内容。若网页为
	// 某搜索引擎的搜索结果页面，则进一步解析出搜索结果地址清单。
	ParseSourceCode(url, sourceCode string) (string, []SearchResult)
}

type SearchResult struct {
	Url  string
	Date time.Time
}

type searchResultExtractor struct {
	mainCss   string
	urlCss    string
	dateCss   string
	timeRe    string
	timeParse string
}

func newSearchResultExtractor(mainCss, urlCss, dateCss, timeRe, timeParse string) *searchResultExtractor {
	return &searchResultExtractor{
		mainCss:   mainCss,
		urlCss:    urlCss,
		dateCss:   dateCss,
		timeRe:    timeRe,
		timeParse: timeParse,
	}
}

func (e *searchResultExtractor) getResults(dom *goquery.Document) []SearchResult {
	res := make([]SearchResult, 0)
	dom.Find(e.mainCss).Each(func(i int, s *goquery.Selection) {
		url, _ := s.Find(e.urlCss).Attr("href")
		dateStr := s.Find(e.dateCss).Text()
		reg := regexp.MustCompile(e.timeRe)
		dateStr = reg.FindString(dateStr)
		var date time.Time
		if dateStr == "" {
			date = time.Now()
		} else {
			date, _ = time.Parse(e.timeParse, dateStr)
		}
		res = append(res, SearchResult{
			Url:  url,
			Date: date,
		})

	})
	sort.Slice(res, func(i, j int) bool {
		return res[i].Date.Before(res[j].Date)
	})
	return res
}

var extractorMap = map[string]*searchResultExtractor{
	"https://cn.bing.com/search": newSearchResultExtractor(
		"li.b_algo",
		".b_title a",
		".b_caption p",
		`\d{4}-\d{1,2}-\d{1,2}`,
		"2006-1-2"),
	"https://www.baidu.com/s": newSearchResultExtractor(
		"div.result.c-container.new-pmd",
		"h3.t a",
		".c-abstract span",
		`\d{4}年\d{1,2}月\d{1,2}日`,
		"2006年1月2日"),
}

// var tags = []string{"p", "li", "h5", "h4", "h3", "h2", "h1"}
// var tagsCombine = strings.Join(tags, ",")

// func getPlainText(dom *goquery.Document) (content string, err error) {
// 	var buffer bytes.Buffer
// 	dom.Find(tagsCombine).Each(func(i int, s *goquery.Selection) {
// 		buffer.WriteString(s.Text())
// 	})
// 	content = buffer.String()
// 	content = strings.ReplaceAll(content, "\n", "")
// 	content = strings.ReplaceAll(content, " ", "")
// 	return
// }

type CommonParser struct {
	parser readability.Parser
}

func NewCommonParser() *CommonParser {
	return &CommonParser{
		parser: readability.NewParser(),
	}
}

func (p *CommonParser) getPlainText(r io.Reader, url string) (content string, err error) {
	parsedURL, err := nurl.ParseRequestURI(url)
	if err != nil {
		return
	}
	article, err := p.parser.Parse(r, parsedURL)
	return article.TextContent, err
}

func (p *CommonParser) ParseSourceCode(url, sourceCode string) (content string, searchResults []SearchResult) {
	var err error
	defer func() {
		if err != nil {
			log.Println("ParseSourceCode err:", err)
		}
	}()
	r := strings.NewReader(sourceCode)
	dom, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return
	}
	// content, err = getPlainText(dom)
	// if err != nil {
	// 	return
	// }
	content, err = p.getPlainText(strings.NewReader(sourceCode), url)
	if err != nil {
		return
	}
	for prefix, extractor := range extractorMap {
		if strings.HasPrefix(url, prefix) {
			searchResults = extractor.getResults(dom)
			break
		}
	}
	return
}
