package octopusapi

import (
	"crypto/tls"
	"encoding/json"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
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

type Images []struct {
	Attributes struct {
		AttributePaVariant string `json:"attribute_pa_variant"`
		AttributePaSize    string `json:"attribute_pa_size"`
	} `json:"attributes"`
	AvailabilityHTML  string `json:"availability_html"`
	BackordersAllowed bool   `json:"backorders_allowed"`
	Dimensions        struct {
		Length string `json:"length"`
		Width  string `json:"width"`
		Height string `json:"height"`
	} `json:"dimensions"`
	DimensionsHTML      string  `json:"dimensions_html"`
	DisplayPrice        float64 `json:"display_price"`
	DisplayRegularPrice float64 `json:"display_regular_price"`
	Image               struct {
		Title                string `json:"title"`
		Caption              string `json:"caption"`
		URL                  string `json:"url"`
		Alt                  string `json:"alt"`
		Src                  string `json:"src"`
		Srcset               bool   `json:"srcset"`
		Sizes                string `json:"sizes"`
		FullSrc              string `json:"full_src"`
		FullSrcW             int    `json:"full_src_w"`
		FullSrcH             int    `json:"full_src_h"`
		GalleryThumbnailSrc  string `json:"gallery_thumbnail_src"`
		GalleryThumbnailSrcW int    `json:"gallery_thumbnail_src_w"`
		GalleryThumbnailSrcH int    `json:"gallery_thumbnail_src_h"`
		ThumbSrc             string `json:"thumb_src"`
		ThumbSrcW            int    `json:"thumb_src_w"`
		ThumbSrcH            int    `json:"thumb_src_h"`
		SrcW                 int    `json:"src_w"`
		SrcH                 int    `json:"src_h"`
	} `json:"image"`
	ImageID              int    `json:"image_id"`
	IsDownloadable       bool   `json:"is_downloadable"`
	IsInStock            bool   `json:"is_in_stock"`
	IsPurchasable        bool   `json:"is_purchasable"`
	IsSoldIndividually   string `json:"is_sold_individually"`
	IsVirtual            bool   `json:"is_virtual"`
	MaxQty               string `json:"max_qty"`
	MinQty               int    `json:"min_qty"`
	PriceHTML            string `json:"price_html"`
	Sku                  string `json:"sku"`
	VariationDescription string `json:"variation_description"`
	VariationID          int    `json:"variation_id"`
	VariationIsActive    bool   `json:"variation_is_active"`
	VariationIsVisible   bool   `json:"variation_is_visible"`
	Weight               string `json:"weight"`
	WeightHTML           string `json:"weight_html"`
}

var defaultHeaders = map[string]string{
	"Accept":     "application/json, text/plain, */*",
	"User-Agent": "User-AgentMozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.45",
}

const site = "https://cuddlyoctopus.com/"

var reProductURL = regexp.MustCompile(site + `product/[^/]+`)
var reProductDetails = regexp.MustCompile(`<script type="application/ld\+json">([\s\S]*?)</script>`) //1=JSON with details
var reImageData = regexp.MustCompile(`\[{&[^"]+`)

func GetProductByURL(URL string) (*Product, error) {
	if !reProductURL.MatchString(URL) {
		return nil, ErrURLParseFailed
	}

	htmlData, err := get(URL)
	if err != nil {
		return nil, err
	}

	matchedProductDetails := reProductDetails.FindSubmatch(htmlData)
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

	matchedImageData := string(reImageData.Find(htmlData))
	matchedImageData = html.UnescapeString(matchedImageData)

	imgData := Images{}
	err = json.Unmarshal([]byte(matchedImageData), &imgData)
	if err != nil {
		return nil, err
	}

	for _, image := range imgData {
		if image.Attributes.AttributePaVariant != "r18" {
			continue
		}
		details.NSFWImage = image.Image.Src
		break
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
