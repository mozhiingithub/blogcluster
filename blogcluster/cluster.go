package blogcluster

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/hbollon/go-edlib"
)

type Cluster struct {
	pageSlices [][]*Page
	c2cluster  chan *Page
	mutex      sync.Mutex
	Dates      map[string]time.Time
	C2Frontend chan []SearchResult
}

func NewCluster(c2cluster chan *Page) *Cluster {
	return &Cluster{
		pageSlices: make([][]*Page, 0),
		c2cluster:  c2cluster,
		mutex:      sync.Mutex{},
		Dates:      make(map[string]time.Time),
		C2Frontend: make(chan []SearchResult, 5),
	}
}

func (c *Cluster) addNewPageGroup(page *Page) {
	c.pageSlices = append(c.pageSlices, []*Page{page})
}

func (c *Cluster) addPageToGroup(page *Page, groupIndex int) {
	c.pageSlices[groupIndex] = append(c.pageSlices[groupIndex], page)
	sort.Slice(c.pageSlices[groupIndex], func(i, j int) bool {
		iDate, _ := c.getPageDate(c.pageSlices[groupIndex][i].Url)
		jDate, _ := c.getPageDate(c.pageSlices[groupIndex][j].Url)
		return iDate.Before(jDate)
	})
}

func (c *Cluster) getPageDate(url string) (time.Time, bool) {
	if t, ok := c.Dates[url]; ok {
		return t, ok
	} else {
		return time.Now(), false
	}
}

func (c *Cluster) addNewPages() {
	for {
		page := <-c.c2cluster
		c.mutex.Lock()
		if page.SearchResults != nil {
			for _, v := range page.SearchResults {
				c.Dates[v.Url] = v.Date
			}
			c.C2Frontend <- page.SearchResults
		}
		pageSliceLen := len(c.pageSlices)
		if pageSliceLen == 0 {
			c.addNewPageGroup(page)
			c.mutex.Unlock()
			continue
		}
		maxI := 0
		var maxSim float32 = 0.0
		for i := 0; i < pageSliceLen; i++ {
			sim2I, _ := edlib.StringsSimilarity(
				c.pageSlices[i][0].Content,
				page.Content,
				edlib.Levenshtein)
			// log.Println("p1:", c.pageSlices[i][0].Url)
			// log.Println("p2:", page.Url)
			// log.Println("sim:", sim2I)
			// log.Println("p1:", c.pageSlices[i][0].Content)
			// log.Println("p2:", page.Content)
			if sim2I > maxSim {
				maxSim = sim2I
				maxI = i
			}
		}
		if maxSim >= 0.5 {
			c.addPageToGroup(page, maxI)
		} else {
			c.addNewPageGroup(page)
		}
		c.mutex.Unlock()
	}
}

func (c *Cluster) String() (res string) {
	var buffer bytes.Buffer
	c.mutex.Lock()
	for i := 0; i < len(c.pageSlices); i++ {
		buffer.WriteString(fmt.Sprintf("Group %d:\n", i+1))
		for j := 0; j < len(c.pageSlices[i]); j++ {
			page := c.pageSlices[i][j]
			pageDate, ok := c.getPageDate(page.Url)
			var dateStr string
			if ok {
				dateStr = pageDate.String()
			} else {
				dateStr = "Unknown"
			}
			pageStr := fmt.Sprintf("title:%s\nurl:%s\ndate:%s\n", page.Title, page.Url, dateStr)
			buffer.WriteString(pageStr)
			buffer.WriteString("-------------------------------------------------\n")
		}
		buffer.WriteString("\n\n")
	}
	c.mutex.Unlock()
	res = buffer.String()
	return
}

func (c *Cluster) Run() {
	c.addNewPages()
}
