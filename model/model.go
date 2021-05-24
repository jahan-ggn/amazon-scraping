package model

import (
	"time"
)

type Product struct {
	ProductTitle        string `json:"name" bson:"name"`
	ProductDescription  string `json:"description" bson:"description"`
	ProductImageURL     string `json:"imageURL" bson:"imageURL"`
	ProductTotalReviews string `json:"totalReviews" bson:"totalReviews"`
	ProductPrice        string `json:"price" bson:"price"`
}
type ProductInformation struct {
	ProductURL string    `json:"url" bson:"url"`
	Product    Product   `json:"product"`
	Timestamp  time.Time `bson:"timestamp,omitempty"`
}
