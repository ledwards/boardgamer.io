package main

import (
  "net/http"
  "io/ioutil"
  "github.com/martini-contrib/encoder"
  "github.com/codegangsta/martini"
  "github.com/codegangsta/martini-contrib/binding"
  "github.com/codegangsta/martini-contrib/render"
  "github.com/moovweb/gokogiri"
  "labix.org/v2/mgo"
)

type Game struct {
  ID string           `json:"id" form:"id"`
  MinPlayers string   `json:"minPlayers" form:"min-players"`
  MaxPlayers string   `json:"maxPlayers" form:"max-players"`
  Name string         `json:"name" form:"name"`
  ImageURL string     `json:"imageURL" form:"image-url"`
  Description string  `json:"description" form:"description"`
}

func DB() martini.Handler {
    session, err := mgo.Dial("mongodb://localhost")
    if err != nil {
        panic(err)
    }

    return func(c martini.Context) {
        s := session.Clone()
        c.Map(s.DB("boardgamer"))
        defer s.Close()
        c.Next()
    }
}

func main() {
  m := martini.New()
  m.Use(martini.Recovery())
  m.Use(martini.Logger())
  m.Use(render.Renderer())
  m.Use(JSONEncoder)
  m.Use(DB())

  r := martini.NewRouter()
  r.Get("/games/:id", GetGame)

  r.Get("/games", func(r render.Render, db *mgo.Database) {
      r.HTML(200, "index", GetGames(db))
  })

  r.Post("/games", binding.Form(Game{}), func(game Game, r render.Render, db *mgo.Database) {
      db.C("games").Insert(game)
      r.HTML(200, "index", GetGames(db))
  })

  m.Action(r.Handle)
  m.Run()
}

func GetGames(db *mgo.Database) []Game {
    var games []Game
    db.C("games").Find(nil).All(&games)
    return games
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

    names, _ := doc.Search("items//item//name[@type='primary']")
    name := names[0].Attribute("value").Value()

    descs, _ := doc.Search("items//item//description")
    desc := descs[0].Content()

    mins, _ := doc.Search("items//item//minplayers")
    min := mins[0].Attribute("value").Value()

    maxes, _ := doc.Search("items//item//maxplayers")
    max := maxes[0].Attribute("value").Value()

    images, _ := doc.Search("items//item//image")
    image := images[0].Content()

    result := &Game{ Description: desc, ID: id, MinPlayers: min, MaxPlayers: max, Name: name, ImageURL: image }
    return http.StatusOK, string(encoder.Must(enc.Encode(result)))
  }
}

func JSONEncoder(c martini.Context, w http.ResponseWriter, r *http.Request) {
  c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
  w.Header().Set("Content-Type", "application/json")
}
