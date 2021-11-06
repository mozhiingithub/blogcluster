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
}
