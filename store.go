package guide

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
)

type store interface {
	GetGuide(int64) (Guide, error) //TODO change return to (Guide,error) should allow the user to change the value directly, should use method
	CreateGuide(guide Guide) (int64, error)
	UpdateGuide(guide Guide) error
	GetAllGuides() []Guide
}
type memoryStore struct {
	Guides       map[int64]Guide
	nextGuideKey int64
	nextPoiKey   int64
}

func OpenMemoryStore() memoryStore {
	ms := memoryStore{
		Guides:       map[int64]Guide{},
		nextGuideKey: 1,
		nextPoiKey:   1,
	}
	return ms
}

func (s *memoryStore) GetGuide(id int64) (Guide, error) {
	guide, ok := s.Guides[id]
	if ok {
		return guide, nil
	}
	return Guide{}, errors.New("guide not found")
}

func (s *memoryStore) CreateGuide(g Guide) (int64, error) {
	g.Id = s.nextGuideKey
	g.Pois = &[]pointOfInterest{}
	s.Guides[g.Id] = g
	s.nextGuideKey++
	return g.Id, nil
}

func (s *memoryStore) UpdateGuide(g Guide) error {
	if g.Id == 0 {
		return errors.New("must set the id of the guide")
	}
	if _, ok := s.Guides[g.Id]; !ok {
		return errors.New("cannot update guide does not exist")
	}
	s.Guides[g.Id] = g
	return nil
}

// GetAllGuides returns a []Guide of all the stored guides
func (s memoryStore) GetAllGuides() []Guide {
	allGuides := make([]Guide, 0, len(s.Guides))
	for _, h := range s.Guides {
		allGuides = append(allGuides, h)
	}
	return allGuides
}

func (s *memoryStore) CreatePoi(poi pointOfInterest) (int64, error) {
	g, ok := s.Guides[poi.GuideID]
	if !ok {
		return 0, errors.New("guide not found")
	}
	poi.Id = s.nextPoiKey
	*g.Pois = append(*g.Pois, poi)
	s.nextPoiKey++
	return poi.Id, nil
}

func (s *memoryStore) UpdatePoi(poi pointOfInterest) error {
	g, ok := s.Guides[poi.GuideID]
	if !ok {
		return errors.New("guide not found")
	}
	found := false
	for i, _ := range *g.Pois {
		if (*g.Pois)[i].Id == poi.Id {
			found = true
			(*g.Pois)[i].Name = poi.Name
			(*g.Pois)[i].Description = poi.Description
			(*g.Pois)[i].Coordinate = poi.Coordinate
		}
	}

	if !found {
		return errors.New("poi not found")
	}
	return nil
}

func (s *memoryStore) GetPoi(id int64) (pointOfInterest, error) {
	for _, g := range s.Guides {
		for _, p := range *g.Pois {
			if p.Id == id {
				return p, nil
			}
		}
	}
	return pointOfInterest{}, errors.New("poi not found")
}

func (s *memoryStore) GetAllPois(guideId int64) []pointOfInterest {
	g, ok := s.Guides[guideId]
	if !ok {
		return []pointOfInterest{}
	}
	return *g.Pois
}

type sqliteStore struct {
	db *sql.DB
}

func OpenSQLiteStore(dbPath string) (sqliteStore, error) {
	if dbPath == "" {
		return sqliteStore{}, errors.New("db source cannot be empty")
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return sqliteStore{}, err
	}

	for _, stmt := range []string{pragmaWALEnabled, pragma500BusyTimeout, pragma500BusyTimeout} {
		_, err = db.Exec(stmt, nil)
		if err != nil {
			return sqliteStore{}, err
		}
	}

	_, err = db.Exec(createGuideTable)
	if err != nil {
		return sqliteStore{}, err
	}

	_, err = db.Exec(createPoiTable)
	if err != nil {
		return sqliteStore{}, err
	}

	store := sqliteStore{
		db: db,
	}
	return store, nil
}

func (s *sqliteStore) CreateGuide(g Guide) (int64, error) {
	stmt, err := s.db.Prepare(insertGuide)
	if err != nil {
		return 0, err
	}

	rs, err := stmt.Exec(g.Name, g.Description, g.Coordinate.Latitude, g.Coordinate.Longitude)
	if err != nil {
		return 0, err
	}
	lastInsertID, err := rs.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

func (s *sqliteStore) GetGuide(id int64) (Guide, error) {
	rows, err := s.db.Query(getGuide, id)
	if err != nil {
		return Guide{}, err
	}

	g := Guide{}

	for rows.Next() {
		var (
			name        string
			description string
			latitude    float64
			longitude   float64
		)
		err = rows.Scan(&name, &description, &latitude, &longitude)
		if err != nil {
			return Guide{}, err
		}
		g.Id = id
		g.Name = name
		g.Description = description
		g.Coordinate = Coordinate{Latitude: latitude, Longitude: longitude}
	}

	if err = rows.Err(); err != nil {
		return Guide{}, err
	}

	return g, nil
}

func (s *sqliteStore) UpdateGuide(g Guide) error {
	stmt, err := s.db.Prepare(updateGuide)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(g.Name, g.Description, g.Coordinate.Latitude, g.Coordinate.Longitude, g.Id)
	if err != nil {
		return err
	}
	return nil
}

func (s *sqliteStore) GetAllGuides() []Guide {
	rows, err := s.db.Query(getAllGuides)
	if err != nil {
		return []Guide{}
	}

	guides := make([]Guide, 0)

	for rows.Next() {
		var (
			id          int64
			name        string
			description string
			latitude    float64
			longitude   float64
		)
		err = rows.Scan(&id, &name, &description, &latitude, &longitude)
		if err != nil {
			return []Guide{}
		}
		g := Guide{
			Id:          id,
			Name:        name,
			Description: description,
			Coordinate:  Coordinate{Latitude: latitude, Longitude: longitude},
		}
		guides = append(guides, g)
	}

	if err = rows.Err(); err != nil {
		return []Guide{}
	}

	return guides
}

func (s *sqliteStore) CreatePoi(poi pointOfInterest) (int64, error) {
	stmt, err := s.db.Prepare(insertPoi)
	if err != nil {
		return 0, err
	}

	rs, err := stmt.Exec(poi.Name, poi.Description, poi.Coordinate.Latitude, poi.Coordinate.Longitude, poi.GuideID)
	if err != nil {
		return 0, err
	}
	lastInsertID, err := rs.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

func (s *sqliteStore) UpdatePoi(poi pointOfInterest) error {
	stmt, err := s.db.Prepare(updatePoi)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(poi.Name, poi.Description, poi.Coordinate.Latitude, poi.Coordinate.Longitude, poi.Id)
	if err != nil {
		return err
	}
	return nil
}

func (s *sqliteStore) GetPoi(id int64) (pointOfInterest, error) {
	rows, err := s.db.Query(getPoi, id)
	if err != nil {
		return pointOfInterest{}, err
	}

	poi := pointOfInterest{}
	for rows.Next() {
		var (
			name        string
			description string
			latitude    float64
			longitude   float64
			guideid     int64
		)
		err = rows.Scan(&name, &description, &latitude, &longitude, &guideid)
		if err != nil {
			return pointOfInterest{}, err
		}
		poi.Id = id
		poi.Name = name
		poi.Description = description
		poi.Coordinate = Coordinate{Latitude: latitude, Longitude: longitude}
		poi.GuideID = guideid
	}

	if err = rows.Err(); err != nil {
		return pointOfInterest{}, err
	}

	return poi, nil
}

func (s *sqliteStore) GetAllPois(guideId int64) []pointOfInterest {
	rows, err := s.db.Query(getAllPois, guideId)
	if err != nil {
		return []pointOfInterest{}
	}

	pois := make([]pointOfInterest, 0)

	for rows.Next() {
		var (
			id          int64
			name        string
			description string
			latitude    float64
			longitude   float64
		)
		err = rows.Scan(&id, &name, &description, &latitude, &longitude)
		if err != nil {
			return []pointOfInterest{}
		}
		p := pointOfInterest{
			Id:          id,
			Name:        name,
			Description: description,
			Coordinate:  Coordinate{Latitude: latitude, Longitude: longitude},
			GuideID:     guideId,
		}
		pois = append(pois, p)
	}

	if err = rows.Err(); err != nil {
		return []pointOfInterest{}
	}

	return pois
}

const pragmaWALEnabled = `PRAGMA journal_mode = WAL;`
const pragma500BusyTimeout = `PRAGMA busy_timeout = 5000;`
const pragmaForeignKeysON = `PRAGMA foreign_keys = on;`

const createGuideTable = `
CREATE TABLE IF NOT EXISTS guide(
id INTEGER NOT NULL PRIMARY KEY,
name TEXT  NOT NULL,
description TEXT,
latitude REAL NOT NULL,
longitude REAL NOT NULL);`

const createPoiTable = `
CREATE TABLE IF NOT EXISTS poi(
id INTEGER NOT NULL PRIMARY KEY,
name TEXT  NOT NULL,
description TEXT,
latitude REAL NOT NULL,
longitude REAL NOT NULL,
guideId INTEGER NOT NULL,
FOREIGN KEY(guideId) REFERENCES guide(id));`

const insertGuide = `INSERT INTO guide(name, description, latitude, longitude ) VALUES (?, ?, ?, ?);`

const insertPoi = `INSERT INTO poi(name, description, latitude, longitude, guideId ) VALUES (?, ?, ?, ?, ?);`

const getGuide = `SELECT name, description, latitude, longitude FROM guide WHERE id = ?`

const getPoi = `SELECT name, description, latitude, longitude, guideid FROM poi WHERE Id = ?`

const updateGuide = `UPDATE guide SET name = ?, description = ?, latitude = ?, longitude = ? WHERE id = ?`

const updatePoi = `UPDATE poi SET name = ?, description = ?, latitude = ?, longitude = ? WHERE id = ?`

const getAllGuides = `SELECT id,name, description, latitude, longitude FROM guide`

const getAllPois = `SELECT id, name, description, latitude, longitude FROM poi WHERE guideid = ?`
