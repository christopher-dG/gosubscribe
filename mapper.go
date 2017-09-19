package gosubscribe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	osuURL = "https://osu.ppy.sh/api"
	osuKey = os.Getenv("OSU_API_KEY")
)

// GetMapper gets a mapper from the osu! API and adds it to the database if necessary.
func GetMapper(name string) (*Mapper, error) {
	mapper := new(Mapper)
	url := fmt.Sprintf("%s/get_user?k=%s&u=%s&type=string", osuURL, osuKey, name)
	log.Printf("requesting from: %s\n", strings.Replace(url, osuKey, "[secure]", 1))
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mappers []Mapper
	err = json.Unmarshal(body, &mappers)
	if err != nil {
		return nil, err
	}

	if len(mappers) == 0 {
		return nil, fmt.Errorf("no mapper was found for %s", name)
	}

	mapper = &mappers[0]

	fromDB := new(Mapper)
	DB.Where("id = ?", mapper.ID).First(fromDB)

	if fromDB.ID == 0 {
		// Add a new mapper.
		mapper.Insert()
	} else if fromDB.ID == mapper.ID && fromDB.Username != mapper.Username {
		// Name change.
		fromDB.Update(mapper.Username)
	}

	return mapper, nil
}

// MapperFromDB gets a mapper from the database.
func MapperFromDB(name string) (*Mapper, error) {
	mapper := new(Mapper)
	DB.Where("lower(username) = lower(?)", name).First(mapper)
	if mapper.ID == 0 {
		return mapper, fmt.Errorf("no mapper was found for %s", name)
	}
	return mapper, nil
}

// Insert adds a new mapper to the databse.
func (mapper *Mapper) Insert() {
	DB.Create(mapper)
	maps, err := mapper.GetMaps()
	if err != nil {
		log.Printf("Maps could not be retrieved for %s\n", mapper.Username)
		return
	}

	// Add the mapper's maps to the DB.
	inserted := []*Map{}
	for _, beatmap := range maps {

		if !HasMap(inserted, beatmap) {
			inserted = append(inserted, beatmap)
			beatmap.MapperID = mapper.ID
			DB.Create(&beatmap)
		}
	}
}

// Update updates a mapper's username.
func (mapper *Mapper) Update(newName string) {
	log.Printf("%s has changed their name to %s\n", mapper.Username, newName)
	mapper.Username = newName
	DB.Save(&mapper)
}

// GetMaps gets all maps by the mapper.
func (mapper *Mapper) GetMaps() ([]*Map, error) {
	var maps []Map
	url := fmt.Sprintf("%s/get_beatmaps?k=%s&u=%d", osuURL, osuKey, mapper.ID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &maps)
	if err != nil {
		return nil, err
	}

	log.Printf("retrieved %d maps\n", len(maps))
	// The API doesn't provide mapper ID so we need to fill it ourselves.
	// Convert to pointers at the same time.
	ptrs := []*Map{}
	for i, beatmap := range maps {
		beatmap.MapperID = mapper.ID
		ptrs = append(ptrs, &maps[i])
	}

	return ptrs, nil
}
