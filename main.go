package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/s-rah/onionscan/config"
	"github.com/s-rah/onionscan/deanonymization"
	"github.com/s-rah/onionscan/onionscan"
	"github.com/s-rah/onionscan/report"
	"github.com/s-rah/onionscan/utils"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("    onionscan [flags] hiddenservice | onionscan [flags] --list list\n")
		flag.PrintDefaults()
	}

	torProxyAddress := flag.String("torProxyAddress", "127.0.0.1:9050", "the address of the tor proxy to use")
	simpleReport := flag.Bool("simpleReport", true, "print out a simple report detailing what is wrong and how to fix it, true by default")
	jsonSimpleReport := flag.Bool("jsonSimpleReport", false, "print out a simple report as json, false by default")
	reportFile := flag.String("reportFile", "", "the file destination path for report file - if given, the prefix of the file will be the scanned onion service. If not given, the report will be written to stdout")
	jsonReport := flag.Bool("jsonReport", false, "print out a json report providing a detailed report of the scan.")
	verbose := flag.Bool("verbose", false, "print out a verbose log output of the scan")
	directoryDepth := flag.Int("depth", 100, "depth of directory scan recursion (default: 100)")
	fingerprint := flag.Bool("fingerprint", true, "true disables some deeper scans e.g. directory probing with the aim of just getting a fingerprint of the service.")
	list := flag.String("list", "", "If provided OnionScan will attempt to read from the given list, rather than the provided hidden service")
	timeout := flag.Int("timeout", 120, "read timeout for connecting to onion services")
	batch := flag.Int("batch", 10, "number of onions to scan concurrently")
	dbdir := flag.String("dbdir", "./onionscandb", "The directory where the crawl database will be stored")
	crawlconfigdir := flag.String("crawlconfigdir", "", "A directory where crawl configurations are stored")
	scans := flag.String("scans", "", "a comma-separated list of scans to run e.g. web,tls,... (default: run all)")

	flag.Parse()

	if len(flag.Args()) != 1 && *list == "" {
		flag.Usage()
		os.Exit(1)
	}

	if !*simpleReport && !*jsonReport && !*jsonSimpleReport {
		log.Fatalf("You must set one of --simpleReport or --jsonReport or --jsonSimpleReport")
	}

	proxyStatus := utils.CheckTorProxy(*torProxyAddress)
	if proxyStatus != utils.ProxyStatusOK {
		log.Fatalf("%s, is the --torProxyAddress setting correct?", utils.ProxyStatusMessage(proxyStatus))
	}

	onionsToScan := []string{}
	if *list == "" {
		onionsToScan = append(onionsToScan, flag.Args()[0])
		log.Printf("Starting Scan of %s\n", flag.Args()[0])
	} else {
		content, err := ioutil.ReadFile(*list)
		if err != nil {
			log.Fatalf("Could not read onion file %s\n", *list)
		}
		onions := strings.Split(string(content), "\n")
		for _, onion := range onions[0 : len(onions)-1] {
			onionsToScan = append(onionsToScan, onion)
		}
		log.Printf("Starting Scan of %d onion services\n", len(onionsToScan))
	}
	log.Printf("This might take a few minutes..\n\n")

	onionScan := new(onionscan.OnionScan)

	var scans_list []string
	if *scans != "" {
		scans_list = strings.Split(*scans, ",")
	} else {
		scans_list = onionScan.GetAllActions()
	}

	onionScan.Config = config.Configure(*torProxyAddress, *directoryDepth, *fingerprint, *timeout, *dbdir, scans_list, *crawlconfigdir, *verbose)

	reports := make(chan *report.OnionScanReport)

	count := 0
	if *batch > len(onionsToScan) {
		*batch = len(onionsToScan)
	}

	// Run an initial batch of 100 requests (or less...)
	for count < *batch {
		go onionScan.Scan(onionsToScan[count], reports)
		count++
	}

	received := 0
	for received < len(onionsToScan) {
		scanReport := <-reports

		// After the initial batch, it's one in one out to prevent proxy overload.
		if count < len(onionsToScan) {
			go onionScan.Scan(onionsToScan[count], reports)
			count++
		}

		received++
		if scanReport.TimedOut {
			onionScan.Config.LogError(errors.New(scanReport.HiddenService + " timed out"))
		}

		file := *reportFile
		if file != "" {
			file = scanReport.HiddenService + "." + *reportFile
		}

		if *jsonReport {
			report.GenerateJsonReport(file, deanonymization.ProcessReport(scanReport, onionScan.Config))
		} else if *jsonSimpleReport {
			report.GenerateSimpleReport(file, deanonymization.ProcessReport(scanReport, onionScan.Config), true, 0)
		} else if *simpleReport {
			termWidth, _, err := terminal.GetSize(int(os.Stdin.Fd()))
			if err != nil {
				termWidth = 80
			}
			report.GenerateSimpleReport(file, deanonymization.ProcessReport(scanReport, onionScan.Config), false, termWidth-1)
		}
	}
}
