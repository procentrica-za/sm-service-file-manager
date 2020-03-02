package main

import (
	"github.com/gorilla/mux"
)

type Server struct {
	router *mux.Router
}

type Config struct {
	ListenServePort string
	ResourcesPath   string
	CRUDHost        string
	CRUDPort        string
}

type CardImage struct {
	EntityID string `json:"entityid"`
	FilePath string `json:"filepath"`
	FileName string `json:"filename"`
}

type CardImageBytes struct {
	EntityID   string `json:"entityid"`
	ImageBytes []byte `json:"imagebytes"`
}

type CardBytesBatch struct {
	Images []CardImageBytes `json:"images"`
}

type CardImageBatch struct {
	Images []CardImage `json:"images"`
}

type CardImageRequest struct {
	EntityID string `json:"entityid"`
}

type CardImageBatchRequest struct {
	Cards []CardImageRequest `json:"cards"`
}
