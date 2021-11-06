package blogcluster

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

// brower模块封装chromedp方法，构造浏览器实例。
// 实例内包含历史记录、各记录的纯文本。
// 浏览器实例与聚类模块建立channel通信，
// 不停地发送新的网页，交由聚类模块更新关系列表。

// 网页（标签页）结构体变量。
type Page struct {
	Url           string         // 网页地址
	Title         string         // 网页标题
	Content       string         // 网页源码
	SearchResults []SearchResult // 搜索结果网址列表（非搜索结果页面为空）
}

type Browser struct {
	remoteAddr     string             // 浏览器远程调试地址
	newTabUrl      string             // 浏览器新标签页地址
	ctx            context.Context    // 远程调试context
	cancelFunc     context.CancelFunc // 远程调试关闭函数
	history        map[string]*Page   // 历史记录，key为URL，value为Page结构体
	mutex          sync.Mutex         // history的互斥锁
	updateInterval time.Duration      // 更新当前标签页列表的间隔时长
	C2cluster      chan *Page         // 与聚类模块沟通的通道，用于发送新的页面信息
	parser         Parser             // 解析器，用于解析html源码，生成纯文本或搜索结果列表
}

// 初始化一个浏览器实例
func NewBrowser(remoteAddr string,
	browerBrand string,
	updateInterval time.Duration,
	parser Parser) *Browser {
	if browerBrand != "edge" && browerBrand != "chrome" {
		log.Println("browser brand wrong.")
		return nil
	}
	ctx, cancelFunc := chromedp.NewRemoteAllocator(context.Background(), remoteAddr)

	return &Browser{
		remoteAddr:     remoteAddr,
		ctx:            ctx,
		cancelFunc:     cancelFunc,
		history:        make(map[string]*Page),
		mutex:          sync.Mutex{},
		updateInterval: updateInterval,
		newTabUrl:      browerBrand + "://newtab/",
		C2cluster:      make(chan *Page, 10),
		parser:         parser,
	}
}

// 在指定间隔时长下，检查当前浏览器标签页列表。
// 发现未曾出现过的新页面，启动携程进行并发解析和存储。
func (b *Browser) saveNewPages() {
	for {
		time.Sleep(b.updateInterval)
		ctx, cancelFunc2 := chromedp.NewContext(b.ctx)
		infos, err := chromedp.Targets(ctx)
		if err != nil {
			log.Println("saveNewPages Error1:", err)
			continue
		}
		for _, info := range infos {
			if info.Type == "page" && info.URL != b.newTabUrl {
				b.mutex.Lock()
				if _, ok := b.history[info.URL]; !ok {
					go b.savePageFromInfo(info)
				}
				b.mutex.Unlock()
			}
		}
		cancelFunc2()
	}
}

// 单个网页保存方法。先调用chromedp方法，获取页面的html源码，
// 再调用浏览器的解析器对html源码进行解析，获取纯文本或搜索结果列表。
// 解析完内容后，上锁检查历史记录当中是否已有该网页
// 若无，则添加该网页。
func (b *Browser) savePageFromInfo(info *target.Info) {
	sourceCode, err := b.getHTMLFromTargetID(info.TargetID)
	if err != nil {
		log.Println("savePageFromInfo err1", err)
		return
	}
	content, searchResults := b.parser.ParseSourceCode(info.URL, sourceCode)

	b.mutex.Lock()
	if _, ok := b.history[info.URL]; !ok {
		page := &Page{
			Url:           info.URL,
			Title:         info.Title,
			Content:       content,
			SearchResults: searchResults,
		}
		b.history[page.Url] = page
		b.C2cluster <- page
	}
	b.mutex.Unlock()

}

func (b *Browser) getHTMLFromTargetID(targetID target.ID) (htmlStr string, err error) {
	tabCtx, _ := chromedp.NewContext(b.ctx, chromedp.WithTargetID(targetID))
	err = chromedp.Run(tabCtx, chromedp.OuterHTML("html", &htmlStr, chromedp.ByQuery))
	return
}

func (b *Browser) Run() {
	defer b.cancelFunc()
	b.saveNewPages()
}
