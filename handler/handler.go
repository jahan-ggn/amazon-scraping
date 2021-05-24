package handler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"amazon-scraping/model"

	"github.com/gocolly/colly"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx = context.TODO()
var collection *mongo.Collection
var client *mongo.Client

func Init() {
	var error error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, error = mongo.Connect(ctx, clientOptions)
	if error != nil {
		log.Fatal(error)
	}
}

func LiveAmazonScraper(url string) model.ProductInformation {
	fmt.Println("Scraping Started....")
	var productInfo = model.ProductInformation{
		ProductURL: "NA",
		Product: model.Product{
			ProductTitle:        "NA",
			ProductDescription:  "NA",
			ProductImageURL:     "NA",
			ProductTotalReviews: "NA",
			ProductPrice:        "NA",
		},
	}
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
		productInfo.ProductURL = r.URL.String()
	})

	c.OnHTML("span#productTitle", func(h *colly.HTMLElement) {
		var productName = StandardizeSpaces(h.Text)
		productInfo.Product.ProductTitle = productName
	})

	c.OnHTML("span#priceblock_ourprice", func(h *colly.HTMLElement) {
		var productPrice = h.Text
		productInfo.Product.ProductPrice = productPrice
	})

	c.OnHTML("div#feature-bullets", func(h *colly.HTMLElement) {
		var productDescription = StandardizeSpaces(h.ChildText("span.a-list-item"))
		productInfo.Product.ProductDescription = productDescription
	})

	c.OnHTML("div#imgTagWrapperId", func(h *colly.HTMLElement) {
		var productImageUrl = h.ChildAttr("img", "src")
		productInfo.Product.ProductImageURL = productImageUrl
	})

	c.OnHTML("div.a-row.a-spacing-medium.averageStarRatingNumerical", func(h *colly.HTMLElement) {
		var productTotalReviews = h.ChildText("span.a-size-base.a-color-secondary")
		var productReviewsCount []byte
		for _, character := range productTotalReviews {
			if string(character) >= "0" && string(character) <= "9" {
				productReviewsCount = append(productReviewsCount, byte(character))
			}
		}
		productInfo.Product.ProductTotalReviews = string(productReviewsCount)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.Visit(url)
	productInfo.Timestamp = time.Now()
	fmt.Println("Done with scraping....")
	return productInfo
}

func StandardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func StoreScrapedData(body model.ProductInformation) (*mongo.InsertOneResult, error) {
	collection = client.Database("db_amazon_scrape").Collection("product_information_master")
	qwery, err := collection.InsertOne(ctx, body)
	fmt.Println("Inserting scraped data....")
	if err != nil {
		return nil, err
	}
	return qwery, nil
}

func UpdateScrapeData(body model.ProductInformation) error {
	collection = client.Database("db_amazon_scrape").Collection("product_information_master")
	var t model.ProductInformation
	filter := bson.M{"url": body.ProductURL}
	update := bson.D{
		{"$set", bson.D{{"product", body.Product}}},
		{"$set", bson.D{{"timestamp", body.Timestamp}}},
	}
	fmt.Println("Updating scraped data....")
	err := collection.FindOneAndUpdate(ctx, filter, update).Decode(&t)
	return err
}

func CheckForURLInDB(url string) bool {
	var result model.ProductInformation
	collection = client.Database("db_amazon_scrape").Collection("product_information_master")
	filter := bson.D{primitive.E{Key: "url", Value: url}}
	collection.FindOne(ctx, filter).Decode(&result)
	if result.ProductURL == "" {
		return false
	}
	return true
}
