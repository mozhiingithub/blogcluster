package main

import (
	"blogcluster"
	"flag"
	"time"
)

var remoteAddr, browserBrand string
var updateInterval int

func init() {
	flag.StringVar(&remoteAddr,
		"addr",
		"http://localhost:9222",
		"browser remote debugging url")
	flag.StringVar(&browserBrand,
		"browser",
		"edge",
		"brower brand, like chrome and edge")
	flag.IntVar(&updateInterval,
		"updateInterval",
		3,
		"interval (second) for browser to update the tag list")
	flag.Parse()
}

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
