package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ShopifyProducts is used for storing response of: https://shopify.dev/api/admin-rest/2022-10/resources/product#get-products
type ShopifyProductsCombined []struct {
	Products []Products `json:"products"`
}
type ShopifyProducts struct {
	Products []Products `json:"products"`
}
type Price struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}
type PresentmentPrices struct {
	Price          Price       `json:"price"`
	CompareAtPrice interface{} `json:"compare_at_price"`
}
type Variants struct {
	ID                   int                 `json:"id"`
	ProductID            int                 `json:"product_id"`
	Title                string              `json:"title"`
	Price                string              `json:"price"`
	Sku                  string              `json:"sku"`
	Position             int                 `json:"position"`
	InventoryPolicy      string              `json:"inventory_policy"`
	CompareAtPrice       interface{}         `json:"compare_at_price"`
	FulfillmentService   string              `json:"fulfillment_service"`
	InventoryManagement  string              `json:"inventory_management"`
	Option1              string              `json:"option1"`
	Option2              interface{}         `json:"option2"`
	Option3              interface{}         `json:"option3"`
	CreatedAt            string              `json:"created_at"`
	UpdatedAt            string              `json:"updated_at"`
	Taxable              bool                `json:"taxable"`
	Barcode              string              `json:"barcode"`
	Grams                int                 `json:"grams"`
	ImageID              int                 `json:"image_id"`
	Weight               float64             `json:"weight"`
	WeightUnit           string              `json:"weight_unit"`
	InventoryItemID      int                 `json:"inventory_item_id"`
	InventoryQuantity    int                 `json:"inventory_quantity"`
	OldInventoryQuantity int                 `json:"old_inventory_quantity"`
	PresentmentPrices    []PresentmentPrices `json:"presentment_prices"`
	RequiresShipping     bool                `json:"requires_shipping"`
	AdminGraphqlAPIID    string              `json:"admin_graphql_api_id"`
}
type Options struct {
	ID        int      `json:"id"`
	ProductID int      `json:"product_id"`
	Name      string   `json:"name"`
	Position  int      `json:"position"`
	Values    []string `json:"values"`
}
type Images struct {
	ID                int           `json:"id"`
	ProductID         int           `json:"product_id"`
	Position          int           `json:"position"`
	CreatedAt         string        `json:"created_at"`
	UpdatedAt         string        `json:"updated_at"`
	Alt               interface{}   `json:"alt"`
	Width             int           `json:"width"`
	Height            int           `json:"height"`
	Src               string        `json:"src"`
	VariantIds        []interface{} `json:"variant_ids"`
	AdminGraphqlAPIID string        `json:"admin_graphql_api_id"`
}
type Image struct {
	ID                int           `json:"id"`
	ProductID         int           `json:"product_id"`
	Position          int           `json:"position"`
	CreatedAt         string        `json:"created_at"`
	UpdatedAt         string        `json:"updated_at"`
	Alt               interface{}   `json:"alt"`
	Width             int           `json:"width"`
	Height            int           `json:"height"`
	Src               string        `json:"src"`
	VariantIds        []interface{} `json:"variant_ids"`
	AdminGraphqlAPIID string        `json:"admin_graphql_api_id"`
}
type Products struct {
	ID                int         `json:"id"`
	Title             string      `json:"title"`
	BodyHTML          string      `json:"body_html"`
	Vendor            string      `json:"vendor"`
	ProductType       string      `json:"product_type"`
	CreatedAt         string      `json:"created_at"`
	Handle            string      `json:"handle"`
	UpdatedAt         string      `json:"updated_at"`
	PublishedAt       string      `json:"published_at"`
	TemplateSuffix    interface{} `json:"template_suffix"`
	PublishedScope    string      `json:"published_scope"`
	Tags              string      `json:"tags"`
	AdminGraphqlAPIID string      `json:"admin_graphql_api_id"`
	Variants          []Variants  `json:"variants"`
	Options           []Options   `json:"options"`
	Images            []Images    `json:"images"`
	Image             Image       `json:"image"`
}
type AppCredentials struct {
	ShopName       string
	APIKey         string
	APIAccessToken string
	APISecretKey   string
	UserAgent      string
}

// {"variant":{"image_id":850703190,"option1":"Purple"}}
type ProductVariant struct {
	VariantData Variant `json:"variant"`
}
type Variant struct {
	ImageID int    `json:"image_id"`
	Option1 string `json:"option1"`
}

// {"image":{"src":"http://example.com/rails_logo.gif"}}
type ImageLink struct {
	ImageLinkData ImageLinkStr `json:"image"`
}
type ImageLinkStr struct {
	Src string `json:"src"`
}

func GetCredentials() AppCredentials {
	viper.SetConfigName("credentials") // name of config file (without extension)
	viper.SetConfigType("env")         // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")           // optionally look for config in the working directory
	err := viper.ReadInConfig()        // Find and read the config file
	if err != nil {                    // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	var tempCreds AppCredentials
	tempCreds.APIAccessToken = viper.GetString("API_ACCESS_TOKEN")
	tempCreds.APIKey = viper.GetString("API_KEY")
	tempCreds.APISecretKey = viper.GetString("API_SECRET_KEY")
	tempCreds.ShopName = viper.GetString("SHOP_NAME")
	tempCreds.UserAgent = viper.GetString("USER_AGENT")
	return tempCreds
}

// RetrieveProducts fetchs and saves every single product *that is listed on the store* data to products.json
func RetrieveProducts() {
	var AppCredentials = GetCredentials()
	client := &http.Client{}
	var err error
	var req *http.Request
	var isFirstFetch = true
	var ongoing = true
	var lastID int
	var productDataStorage []ShopifyProducts
	for ongoing {
		if isFirstFetch {
			req, err = http.NewRequest("GET", fmt.Sprintf("https://%s.myshopify.com/admin/api/2022-10/products.json?limit=250", AppCredentials.ShopName), nil)
			if err != nil {
				log.Fatal(err)
			}
			isFirstFetch = false
		} else {
			req, err = http.NewRequest("GET", fmt.Sprintf("https://%s.myshopify.com/admin/api/2022-10/products.json?limit=250&since_id=%d", AppCredentials.ShopName, lastID), nil)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Println(req.URL)
		req.Header.Set("X-Shopify-Access-Token", AppCredentials.APIAccessToken)
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			ongoing = false
		}
		if ongoing {
			defer resp.Body.Close()
			bodyText, err := ioutil.ReadAll(resp.Body)
			// take url from nextFetchLink

			if err != nil {
				log.Fatal(err)
			}
			var tempResponseData ShopifyProducts
			json.Unmarshal(bodyText, &tempResponseData)
			if len(tempResponseData.Products) == 0 {
				ongoing = false
			} else {
				if lastID != (tempResponseData.Products[len(tempResponseData.Products)-1].ID) {
					lastID = (tempResponseData.Products[len(tempResponseData.Products)-1].ID)
					productDataStorage = append(productDataStorage, tempResponseData)
					time.Sleep(time.Second * 1)
				} else {
					ongoing = false
				}
			}
		}
	}
	jsontowrite, err := json.Marshal(productDataStorage)
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile("products.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	if _, err = file.WriteString(string(jsontowrite)); err != nil {
		panic(err)
	}
}
func (c AppCredentials) AddVariantToExisting(imageID int, varName string, productID int) (error, bool) {
	var ok = true
	client := &http.Client{}
	data := ProductVariant{
		VariantData: Variant{
			ImageID: imageID,
			Option1: varName,
		},
	}
	bodyData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s.myshopify.com/admin/api/2022-10/products/%d/variants.json", c.ShopName, productID), strings.NewReader(string(bodyData)))
	if err != nil {
		log.Println(err)
		return err, !ok
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/3.1 (Windows 10)", c.UserAgent))
	req.Header.Set("X-Shopify-Access-Token", c.APIAccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err, !ok
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err, !ok
	}
	fmt.Printf("%s\n", bodyText)
	return err, ok
}
func (c AppCredentials) AddImageToExisting(productID int, imageLink string) {
	client := &http.Client{}
	var data = ImageLink{
		ImageLinkData: ImageLinkStr{
			Src: imageLink,
		},
	}
	bodyData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s.myshopify.com/admin/api/2022-10/products/%d/images.json", c.ShopName, productID), strings.NewReader(string(bodyData)))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/3.1 (Windows 10)", c.UserAgent))
	req.Header.Set("X-Shopify-Access-Token", c.APIAccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", bodyText)
}
func main() {
	var Prods ShopifyProductsCombined
	productData, err := ioutil.ReadFile("products.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(productData, &Prods)
	var credentials = GetCredentials()
	// For every product section
	for i := range Prods {
		// and every products
		for j := range Prods[i].Products {
			// iterate products twice.
			for z := range Prods[i].Products {
				// if both iteration product types are same, and titles arent same, add second iteration to first as a variant.
				if strings.Contains(Prods[i].Products[j].Title, "Pet Premium Jersey") &&
					strings.Contains(Prods[i].Products[z].Title, "Pet Premium Jersey") &&
					(Prods[i].Products[z].Title != Prods[i].Products[j].Title) {
					err, ok := credentials.AddVariantToExisting(Prods[i].Products[z].Image.ID, Prods[i].Products[z].Title, Prods[i].Products[j].ID)
					if !ok {
						log.Fatalln(err)
					}
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}

}
