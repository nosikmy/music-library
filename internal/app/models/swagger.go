package models

type LibraryResponse struct {
	Status  string `json:"status" example:"200"`
	Message string `json:"message" example:"ok"`
	Payload struct {
		Count   int    `json:"count" example:"10"`
		Library []Song `json:"library"`
	}
}

type SongTextResponse struct {
	Status  string `json:"status" example:"200"`
	Message string `json:"message" example:"ok"`
	Payload struct {
		Count int    `json:"count" example:"5"`
		Text  []Song `json:"text"`
	}
}

type AddSongResponse struct {
	Status  string `json:"status" example:"200"`
	Message string `json:"message" example:"ok"`
	Payload struct {
		Id int `json:"id" example:"48"`
	}
}
