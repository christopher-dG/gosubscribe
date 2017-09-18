package gosubscribe

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	osuURL string = "https://osu.ppy.sh/api"
	osuKey string = os.Getenv("OSU_API_KEY")
)

func GetMapper(name string) (Mapper, error) {
	var mapper Mapper
	url := fmt.Sprintf("%s/get_user?k=%s&u=%s&type=string", osuURL, osuKey, name)
	log.Printf("requesting from: %s\n", strings.Replace(url, osuKey, "[secure]", 1))
	resp, err := http.Get(url)
	if err != nil {
		return mapper, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return mapper, err
	}

	var mappers []Mapper
	err = json.Unmarshal(body, &mappers)
	if err != nil {
		return mapper, err
	}

	if len(mappers) == 0 {
		return mapper, errors.New("no mapper was found for " + name)
	}

	mapper = mappers[0]

	var fromDB Mapper
	DB.Where("id = ?", mapper.ID).First(&fromDB)

	if fromDB.ID == 0 {
		mapper.Insert()
	} else if fromDB.ID == mapper.ID && fromDB.Username != mapper.Username {
		fromDB.Update(mapper.Username)
	}

	return mapper, nil
}

func MapperFromDB(name string) (Mapper, error) {
	var mapper Mapper
	DB.Where("lower(username) = lower(?)", name).First(&mapper)
	if mapper.ID == 0 {
		return mapper, errors.New("no mapper was found for " + name)
	} else {
		return mapper, nil
	}
}

func (mapper *Mapper) Insert() {
	DB.Create(&mapper)
	maps, err := mapper.GetMaps()
	if err != nil {
		log.Printf("Maps could not be retrieved for %s\n", mapper.Username)
	}

	inserted := []Map{}
	for _, beatmap := range maps {

		if !contains(inserted, beatmap) {
			inserted = append(inserted, beatmap)
			beatmap.MapperID = mapper.ID
			DB.Create(&beatmap)
		}
	}
}

func (mapper *Mapper) Update(newName string) {
	log.Printf("%s has changed their name to %s\n", mapper.Username, newName)
	mapper.Username = newName
	DB.Save(&mapper)
}

func (mapper *Mapper) GetMaps() ([]Map, error) {
	var maps []Map
	url := fmt.Sprintf("%s/get_beatmaps?k=%s&u=%d", osuURL, osuKey, mapper.ID)
	resp, err := http.Get(url)
	if err != nil {
		return maps, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return maps, err
	}

	err = json.Unmarshal(body, &maps)
	if err != nil {
		return maps, err
	}
	// The API doesn't provide mapper ID so we need to fill it ourselves.
	for _, beatmap := range maps {
		beatmap.MapperID = mapper.ID
	}

	return maps, nil
}

func contains(maps []Map, key Map) bool {
	for _, beatmap := range maps {
		if beatmap.ID == key.ID {
			return true
		}
	}
	return false
}
