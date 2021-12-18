package octopusapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Product details
type Product struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	ID          string `json:"@id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	MainImage   string `json:"image"`
	Sku         int    `json:"sku"`
	Offers      []struct {
		Type               string `json:"@type"`
		Price              string `json:"price"`
		PriceValidUntil    string `json:"priceValidUntil"`
		PriceSpecification struct {
			Price                 string `json:"price"`
			PriceCurrency         string `json:"priceCurrency"`
			ValueAddedTaxIncluded string `json:"valueAddedTaxIncluded"`
		} `json:"priceSpecification"`
		PriceCurrency string `json:"priceCurrency"`
		Availability  string `json:"availability"`
		URL           string `json:"url"`
		Seller        struct {
			Type string `json:"@type"`
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"seller"`
	} `json:"offers"`
	NSFWImage string
}

var defaultHeaders = map[string]string{
	"Accept":     "application/json, text/plain, */*",
	"User-Agent": "User-AgentMozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.45",
}

const site = "https://cuddlyoctopus.com/"

var reProductURL = regexp.MustCompile(site + `product/[^/]+`)
var reProductDetails = regexp.MustCompile(`<script type="application/ld\+json">([\s\S]*?)</script>`) //1=JSON with details
var reImageID = regexp.MustCompile(site + `wp-content/uploads/\d+/\d\d/(\d+)`)                       //1=image ID
var reNSFWImageFallback = regexp.MustCompile(site + `wp-content/uploads/\d+/\d\d/(\d+-\d+)`)         //2=image ID + part - 919-1 -> 920 for NSFW

func GetProductByURL(URL string) (*Product, error) {
	if !reProductURL.MatchString(URL) {
		return nil, ErrURLParseFailed
	}

	htmlString, err := get(URL)
	if err != nil {
		return nil, err
	}

	matchedProductDetails := reProductDetails.FindSubmatch(htmlString)
	if len(matchedProductDetails) == 0 {
		return nil, ErrProductDetailParseFailed
	}

	details := &Product{}
	err = json.Unmarshal(matchedProductDetails[1], details)
	if err != nil {
		return nil, err
	}

	details.Name = html.UnescapeString(details.Name)
	details.Description = html.UnescapeString(details.Description)

	matchedImageID := reImageID.FindStringSubmatch(details.MainImage)
	if len(matchedImageID) < 2 {
		return nil, ErrImageIDParseFailed
	}

	imageID, err := strconv.Atoi(matchedImageID[1])
	if err != nil {
		return nil, err
	}

	details.NSFWImage = strings.Replace(details.MainImage, matchedImageID[1], fmt.Sprint(imageID+1), 1)
	if bytes.Contains(htmlString, []byte(strings.ReplaceAll(details.NSFWImage, "/", `\/`))) {
		return details, nil
	}

	matchedImageURLAlt := reNSFWImageFallback.FindSubmatch(htmlString)
	if len(matchedImageURLAlt) < 2 {
		details.NSFWImage = ""
		return details, nil
	}

	details.NSFWImage = strings.Replace(details.MainImage, string(matchedImageURLAlt[1]), fmt.Sprint(imageID+1), 1)
	if !bytes.Contains(htmlString, []byte(strings.ReplaceAll(details.NSFWImage, "/", `\/`))) {
		details.NSFWImage = ""
	}

	return details, nil
}

func get(URL string) ([]byte, error) {

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression:  true,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			IdleConnTimeout:     5 * time.Second,
		},
		Timeout: 5 * time.Minute,
	}

	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range defaultHeaders {
		req.Header.Set(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		if err != io.ErrUnexpectedEOF {
			return nil, err
		}
	}

	return body, nil

}
