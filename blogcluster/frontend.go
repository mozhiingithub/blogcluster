package blogcluster

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

type TextFrontEndPrinter struct {
	searchResultDir   string
	clusterResultPath string
	cluster           *Cluster
}

func NewTextFrontEndPrinter(
	searchResultDir,
	clusterResultPath string,
	cluster *Cluster) *TextFrontEndPrinter {
	return &TextFrontEndPrinter{
		searchResultDir:   searchResultDir,
		clusterResultPath: clusterResultPath,
		cluster:           cluster,
	}
}

func (p *TextFrontEndPrinter) printSearchResults() {
	for {
		searchResults := <-p.cluster.C2Frontend
		var buffer bytes.Buffer
		for _, v := range searchResults {
			buffer.WriteString(v.Url + "\n")
		}
		savePath := path.Join(p.searchResultDir, fmt.Sprintf("%d.txt", time.Now().Unix()))
		f, _ := os.Create(savePath)
		f.WriteString(buffer.String())
		f.Close()
	}
}

func (p *TextFrontEndPrinter) printCluster() {
	for {
		time.Sleep(3 * time.Second)
		clusterStr := p.cluster.String()
		f, err := os.OpenFile(p.clusterResultPath, os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatal(err)
		}
		writer := bufio.NewWriter(f)
		writer.WriteString(clusterStr)
		writer.Flush()
		f.Close()
	}
}

func (p *TextFrontEndPrinter) Run() {
	go p.printSearchResults()
	go p.printCluster()
}
