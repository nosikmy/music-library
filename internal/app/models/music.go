package models

import (
	"time"
)

type Song struct {
	Id          int       `json:"id" db:"id" example:"458"`
	Name        string    `json:"name" db:"name" example:"Supermassive Black Hole"`
	ReleaseDate time.Time `json:"releaseDate" db:"release_date" example:"16.07.2006"`
	Link        string    `json:"link" db:"link" example:"https://www.youtube.com/watch?v=Xsp3_a-PMTw"`
	Groups      []Group   `json:"groups" db:"groups"`
}

type Group struct {
	Id   int    `json:"groupId" db:"group_id" example:"26"`
	Name string `json:"groupName" db:"group_name" example:"Muse"`
}

type Verse struct {
	Id   int    `json:"verseId" db:"id" example:"89"`
	Text string `json:"text" db:"text" example:"Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?"`
}

type SongDBFormat struct {
	Id          int       `db:"id"`
	Name        string    `db:"name"`
	ReleaseDate time.Time `db:"release_date"`
	Link        string    `db:"link"`
	GroupId     int       `db:"group_id"`
	GroupName   string    `db:"group_name"`
}
