package main

import (
	"amazon-scraping/handler"
	"amazon-scraping/model"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func main() {
	fmt.Println("Application Started....")
	handler.Init()
	Crawl()
}

func Crawl() {
	app := fiber.New()

	app.Post("/scrape-url", func(c *fiber.Ctx) error {
		type Request struct {
			URL string `json:"url"`
		}
		var requestBody Request
		parsingError := c.BodyParser(&requestBody)

		if parsingError != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   parsingError.Error(),
			})
		}
		var scrapedData = handler.LiveAmazonScraper(requestBody.URL)
		if scrapedData.ProductURL == "NA" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "sorry request exceeded at amazon",
			})
		}
		response, error := CallApi(scrapedData)
		if error != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   error.Error(),
			})
		}
		defer response.Body.Close()
		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println()
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Data stored successfully",
			"data":    json.RawMessage(responseBody),
		})
	})

	app.Post("/store-scrape-data", func(c *fiber.Ctx) error {
		var requestBody model.ProductInformation
		parsingError := c.BodyParser(&requestBody)
		if parsingError != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   parsingError.Error(),
			})
		}
		isFound := handler.CheckForURLInDB(requestBody.ProductURL)
		if !isFound {
			_, error := handler.StoreScrapedData(requestBody)
			if error != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": error.Error(),
				})
			}
			return c.Status(fiber.StatusOK).JSON(requestBody)
		}
		error := handler.UpdateScrapeData(requestBody)
		if error != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": error.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(requestBody)
	})

	app.Listen(":3000")
}

func CallApi(data model.ProductInformation) (*http.Response, error) {
	jsonRequest, _ := json.Marshal(data)
	response, error := http.Post("http://localhost:3000/store-scrape-data", "application/json; charset=utf-8", bytes.NewBuffer(jsonRequest))
	if error != nil {
		return nil, error
	}
	return response, nil
}
