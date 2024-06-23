package book

import (
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/author"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/genre"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
	Id     primitive.ObjectID `json:"id"`
	Name   string             `json:"name"`
	Author *author.Author     `json:"author,omitempty"`
	Genres *[]genre.Genre     `json:"genres,omitempty"`
	Pages  int                `json:"pages"`
	Year   int                `json:"year"`
	// Name of book file
	FilePath string `json:"file"`
	// Name of cover file
	CoverPath string `json:"cover"`
}

type InsertBookQuery struct {
	// Book's name. Must be unique
	Name string `form:"name" validate:"required,min=4,max=32"`
	// primitive object id to author of book
	AuthorId primitive.ObjectID `form:"authorid" validate:"required,primitiveid"`
	// Array of genre's Id's (must be primitive object id)
	GenresId []primitive.ObjectID `form:"genres" validate:"required"`
	Year     int                  `form:"year" validate:"required,gte=1400,lte=2100"`
	// Total number of pages in pdf file
	Pages int `form:"pages" validate:"required,lte=5000"`
	// Must be a .pdf file to book
	File string `form:"file" format:"binary" validate:"required"`
	// Must be an image file to book cover
	Cover string `form:"cover" format:"binary" validate:"required"`
}
