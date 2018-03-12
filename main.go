package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func warningForPrice(oldPrice, newPrice float64, firstRun, wasStable bool) bool {

	if firstRun {
		priceWarning1("  PRICE FETCHED", newPrice)
		return false
	}

	if newPrice == oldPrice {
		if wasStable {
			fmt.Print("\033[1A") // move cursor one line up
			fmt.Print("\033[K")  // delete till end of line
		}
		priceWarning1("PRICE IS STABLE", newPrice)
		return true
	}

	if newPrice < oldPrice {
		priceWarning2("PRICE DECREASED", oldPrice, newPrice)
	} else {
		priceWarning2("PRICE INCREASED", oldPrice, newPrice)
	}

	return false
}

func priceWarning1(header string, price float64) {
	log.Printf("%s (at %.2f)\n", header, price)
}

func priceWarning2(header string, from, to float64) {
	log.Printf("%s (from %.2f to %.2f)\n", header, from, to)
}

func downloadFile(url, filepath string) error {
	fmt.Println("Downloading:", filepath)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func getDocument(url string) *goquery.Document {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func getPrice(url string) float64 {
	doc := getDocument(url)

	priceLine := doc.Find("#priceblock_ourprice").Contents().Text()

	priceStr := strings.Replace(priceLine, "EUR ", "", 1)
	priceStr = strings.Replace(priceStr, ",", ".", 1)
	priceStr = strings.Replace(priceStr, "$", "", 1)

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		log.Fatal(err)
	}

	return price
}

func main() {
	url := flag.String("url", "", "Amazon URL to track")
	every := flag.Duration("every", time.Duration(30*time.Minute), "Amount of time to wait between price checks")

	flag.Parse()

	oldPrice := 0.0
	stable := false
	runs := 0
	for {
		newPrice := getPrice(*url)
		stable = warningForPrice(oldPrice, newPrice, runs == 0, stable)
		oldPrice = newPrice

		runs++
		time.Sleep(*every)
	}

}
