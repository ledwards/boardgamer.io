package main

import (
  "net/http"
  "io/ioutil"
  "github.com/martini-contrib/encoder"
  "github.com/codegangsta/martini"
  "github.com/moovweb/gokogiri"
)

type DB interface {
    Get(id int) *Game
    Find(query string) []*Game
}

type Game struct {
  ID string           `json:"id"`
  MinPlayers string   `json:"minPlayers"`
  MaxPlayers string   `json:"maxPlayers"`
  Title string        `json:"title"`
  ImageURL string     `json:"imageURL"`
  Description string  `json:"description`
}

func main() {
  m := martini.New()
  m.Use(martini.Recovery())
  m.Use(martini.Logger())
  m.Use(JSONEncoder)

  r := martini.NewRouter()
  r.Get("/games/:id", GetGame)

  m.Action(r.Handle)
  m.Run()
}

func GetGame(enc encoder.Encoder, params martini.Params) (int, string) {
  response, err := http.Get("http://www.boardgamegeek.com/xmlapi2/thing?id=" + params["id"])

  if err != nil {
    return http.StatusNotFound, "Error"
  } else {
    defer response.Body.Close()
    body, _ := ioutil.ReadAll(response.Body)
    doc, _ := gokogiri.ParseXml(body)
    defer doc.Free()

    ids, _ := doc.Search("items//item")
    id := ids[0].Attribute("id").Value()

    titles, _ := doc.Search("items//item//name[@type='primary']")
    title := titles[0].Attribute("value").Value()

    descs, _ := doc.Search("items//item//description")
    desc := descs[0].Content()

    mins, _ := doc.Search("items//item//minplayers")
    min := mins[0].Attribute("value").Value()

    maxes, _ := doc.Search("items//item//maxplayers")
    max := maxes[0].Attribute("value").Value()

    images, _ := doc.Search("items//item//image")
    image := images[0].Content()

    result := &Game{ Description: desc, ID: id, MinPlayers: min, MaxPlayers: max, Title: title, ImageURL: image }
    return http.StatusOK, string(encoder.Must(enc.Encode(result)))
  }
}

func JSONEncoder(c martini.Context, w http.ResponseWriter, r *http.Request) {
  c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
  w.Header().Set("Content-Type", "application/json")
}
