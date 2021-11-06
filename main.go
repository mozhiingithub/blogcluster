package main

import (
	"blogcluster"
	"flag"
	"time"
)

var remoteAddr, browserBrand, searchResultDir, clusterResultPath string
var updateInterval int

func init() {
	flag.StringVar(&remoteAddr,
		"addr",
		"http://localhost:9222",
		"Browser remote debugging url.")
	flag.StringVar(&browserBrand,
		"browser",
		"edge",
		"Brower brand, like chrome and edge.")
	flag.IntVar(&updateInterval,
		"updateInterval",
		3,
		"Interval (second) for browser to update the tag list.")
	flag.StringVar(&searchResultDir,
		"searchResultDir",
		"searchResults",
		"Dir for saving the URLs of the search results extracted from the search engine pages.")
	flag.StringVar(&clusterResultPath,
		"result",
		"res.txt",
		"Text file to save clustering results.")
	flag.Parse()
}

func main() {
	parser := blogcluster.NewCommonParser()
	browser := blogcluster.NewBrowser(
		remoteAddr,
		browserBrand,
		time.Duration(updateInterval)*time.Second,
		parser,
	)
	cluster := blogcluster.NewCluster(browser.C2cluster)
	printer := blogcluster.NewTextFrontEndPrinter(
		searchResultDir,
		clusterResultPath,
		cluster,
	)
	go browser.Run()
	go cluster.Run()
	go printer.Run()
	select {}
}
