package main

import (
	"blogcluster"
	"time"
)

func main() {
	parser := blogcluster.NewCommonParser()
	browser := blogcluster.NewBrowser(
		"http://localhost:9222",
		"edge",
		3*time.Second,
		parser,
	)
	cluster := blogcluster.NewCluster(browser.C2cluster)
	printer := blogcluster.NewTextFrontEndPrinter(
		"searchResults",
		"res.txt",
		cluster,
	)
	go browser.Run()
	go cluster.Run()
	go printer.Run()
	select {}

	// for {
	// 	time.Sleep(3 * time.Second)
	// 	// page := <-browser.C2cluster
	// 	// saveName := fmt.Sprintf("%d", time.Now().Unix())
	// 	// plainTextSavePath := path.Join(saveDir, fmt.Sprintf("%s.txt", saveName))
	// 	// searchResultSavePath := path.Join(saveDir, fmt.Sprintf("%s_search.txt", saveName))
	// 	// log.Printf("%s->%s\n", page.Title, saveName)
	// 	// f, _ := os.Create(plainTextSavePath)
	// 	// f.WriteString(page.Content)
	// 	// f.Close()
	// 	// if page.SearchResults != nil {
	// 	// 	var buffer bytes.Buffer
	// 	// 	for _, v := range page.SearchResults {
	// 	// 		buffer.WriteString(v.Url + "\n")
	// 	// 	}
	// 	// 	f, _ := os.Create(searchResultSavePath)
	// 	// 	f.WriteString(buffer.String())
	// 	// 	f.Close()
	// 	// }
	// 	clusterStr := cluster.String()
	// 	f, err := os.OpenFile("res.txt", os.O_WRONLY|os.O_TRUNC, 0666)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	writer := bufio.NewWriter(f)
	// 	writer.WriteString(clusterStr)
	// 	writer.Flush()
	// 	f.Close()
	// }
}
