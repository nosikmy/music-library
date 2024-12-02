package models

type Response struct {
	Status  int    `json:"status" example:"200"`
	Message string `json:"message" example:"ok"`
	Payload any    `json:"payload" swaggertype:"string" example:"null"`
}

type ApiMusicRequest struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}

type ApiMusicResponse struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}
