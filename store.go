package guide

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	CreateGuide(*guide) error
	GetGuidebyID(int64) (*guide, error)
	UpdateGuide(*guide) error
	GetAllGuides() []guide

	GetPoi(int64) (pointOfInterest, error)
	CreatePoi(string, int64, ...poiOption) (*pointOfInterest, error)
	UpdatePoi(*pointOfInterest) error
	GetAllPois(int64) []pointOfInterest
}

// todo should be pointers and Posgres
//
//	type Storage interface {
//		CreateMatch(*match) error
//		//DeleteMatch(int)error
//		UpdateMatch(*match) error
//		GetMatchByID(int64) (*match, error)
//		AddPointsByID(int64, int, int) (*match, error)
//	}
type memoryStore struct {
	Guides       map[int64]guide
	Pois         map[int64]pointOfInterest
	Users        map[int64]user
	NextGuideKey int64
	NextPoiKey   int64
	NextUserKey  int64
}

func OpenMemoryStore() memoryStore {
	ms := memoryStore{
		Guides:       map[int64]guide{},
		Pois:         map[int64]pointOfInterest{},
		Users:        map[int64]user{},
		NextGuideKey: 1,
		NextPoiKey:   1,
		NextUserKey:  1,
	}
	return ms
}

func (s *memoryStore) GetGuide(id int64) (guide, error) {
	g, ok := s.Guides[id]
	if ok {
		return g, nil
	}
	return guide{}, errors.New("g not found")
}

func (s *memoryStore) CreateGuide(name string, opts ...guideOption) (*guide, error) {
	g, err := NewGuide(name, opts...)
	if err != nil {
		return nil, err
	}
	g.Id = s.NextGuideKey
	s.Guides[g.Id] = g
	s.NextGuideKey++
	return &g, nil
}

func (s *memoryStore) UpdateGuide(g *guide) error {
	if g.Id == 0 {
		return errors.New("must set the Id of the guide")
	}
	if _, ok := s.Guides[g.Id]; !ok {
		return errors.New("cannot update guide does not exist")
	}
	s.Guides[g.Id] = *g
	return nil
}

// GetAllGuides returns a []guide of all the stored guides
func (s memoryStore) GetAllGuides() []guide {
	allGuides := make([]guide, 0, len(s.Guides))
	for _, h := range s.Guides {
		allGuides = append(allGuides, h)
	}
	return allGuides
}

func (s *memoryStore) CreatePoi(name string, guideID int64, opts ...poiOption) (*pointOfInterest, error) {
	poi, err := newPointOfInterest(name, guideID, opts...)
	if err != nil {
		return nil, err
	}
	_, ok := s.Guides[poi.GuideID]
	if !ok {
		return nil, errors.New("guide not found")
	}
	poi.Id = s.NextPoiKey
	s.Pois[poi.Id] = poi
	s.NextPoiKey++
	return &poi, nil
}

func (s *memoryStore) UpdatePoi(poi *pointOfInterest) error {
	s.Pois[poi.Id] = *poi //todo validate poi
	return nil
}

func (s *memoryStore) GetPoi(id int64) (pointOfInterest, error) {
	poi, ok := s.Pois[id]
	if ok {
		return poi, nil
	}
	return pointOfInterest{}, errors.New("poi not found")
}

func (s *memoryStore) GetAllPois(guideId int64) []pointOfInterest {
	_, ok := s.Guides[guideId]
	if !ok {
		return []pointOfInterest{}
	}
	//todo why is this a warning
	//Empty slice declaration using a literal
	var pois []pointOfInterest
	for _, poi := range s.Pois {
		if poi.GuideID == guideId {
			pois = append(pois, poi)
		}
	}
	return pois
}

type sqliteStore struct {
	db *sql.DB
}

func OpenSQLiteStorage(dbPath string) (Storage, error) {
	if dbPath == "" {
		return &sqliteStore{}, errors.New("db source cannot be empty")
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return &sqliteStore{}, err
	}

	//actually now that I think about it. Is this the migration part of a webapp?
	for _, stmt := range []string{pragmaWALEnabled, pragma500BusyTimeout, pragmaForeignKeysON} {
		_, err = db.Exec(stmt, nil)
		if err != nil {
			return &sqliteStore{}, err
		}
	}

	_, err = db.Exec(createGuideTable)
	if err != nil {
		return &sqliteStore{}, err
	}

	_, err = db.Exec(createPoiTable)
	if err != nil {
		return &sqliteStore{}, err
	}

	//leaving commented to help with the concept of migration
	//_, err = db.Exec(createUserTable)
	//if err != nil {
	//	return &sqliteStore{}, err
	//}

	store := sqliteStore{
		db: db,
	}
	return &store, nil
}

func (s *sqliteStore) CreateGuide(guide *guide) error {
	stmt, err := s.db.Prepare(insertGuide)
	if err != nil {
		return err
	}
	rs, err := stmt.Exec(guide.Name, guide.Description, guide.Coordinate.Latitude, guide.Coordinate.Longitude)
	if err != nil {
		return err
	}
	lastInsertID, err := rs.LastInsertId()
	if err != nil {
		return err
	}
	guide.Id = lastInsertID
	return nil
}

func (s *sqliteStore) GetGuidebyID(id int64) (*guide, error) {
	var (
		name        string
		description string
		latitude    float64
		longitude   float64
	)
	err := s.db.QueryRow(getGuide, id).Scan(&name, &description, &latitude, &longitude)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		g := guide{
			Id:          id,
			Name:        name,
			Description: description,
			Coordinate: coordinate{
				Latitude:  latitude,
				Longitude: longitude,
			},
			Pois: nil,
		}
		return &g, nil
	}
}

func (s *sqliteStore) UpdateGuide(g *guide) error {
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

func (s *sqliteStore) GetAllGuides() []guide {
	rows, err := s.db.Query(getAllGuides)
	if err != nil {
		return []guide{}
	}

	guides := make([]guide, 0)

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
			return []guide{}
		}
		g := guide{
			Id:          id,
			Name:        name,
			Description: description,
			Coordinate:  coordinate{Latitude: latitude, Longitude: longitude},
		}
		guides = append(guides, g)
	}

	if err = rows.Err(); err != nil {
		return []guide{}
	}

	return guides
}

func (s *sqliteStore) CreatePoi(name string, guideID int64, opts ...poiOption) (*pointOfInterest, error) {
	poi, err := newPointOfInterest(name, guideID, opts...)
	if err != nil {
		return nil, err
	}
	stmt, err := s.db.Prepare(insertPoi)
	if err != nil {
		return nil, err
	}

	rs, err := stmt.Exec(poi.Name, poi.Description, poi.Coordinate.Latitude, poi.Coordinate.Longitude, poi.GuideID)
	if err != nil {
		return nil, err
	}
	lastInsertID, err := rs.LastInsertId()
	if err != nil {
		return nil, err
	}
	poi.Id = lastInsertID
	return &poi, nil
}

func (s *sqliteStore) UpdatePoi(poi *pointOfInterest) error {
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
		poi.Coordinate = coordinate{Latitude: latitude, Longitude: longitude}
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
			Coordinate:  coordinate{Latitude: latitude, Longitude: longitude},
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
Id INTEGER NOT NULL PRIMARY KEY,
name TEXT  NOT NULL,
description TEXT,
latitude REAL NOT NULL,
longitude REAL NOT NULL,
CHECK (name <> ''));`

const createPoiTable = `
CREATE TABLE IF NOT EXISTS poi(
Id INTEGER NOT NULL PRIMARY KEY,
name TEXT  NOT NULL,
description TEXT,
latitude REAL NOT NULL,
longitude REAL NOT NULL,
guideId INTEGER NOT NULL,
FOREIGN KEY(guideId) REFERENCES guide(Id),
CHECK (name <> ''));`

const insertGuide = `INSERT INTO guide(name, description, latitude, longitude ) VALUES (?, ?, ?, ?);`

const insertPoi = `INSERT INTO poi(name, description, latitude, longitude, guideId ) VALUES (?, ?, ?, ?, ?);`

const getGuide = `SELECT name, description, latitude, longitude FROM guide WHERE Id = ?`

const getPoi = `SELECT name, description, latitude, longitude, guideid FROM poi WHERE Id = ?`

const updateGuide = `UPDATE guide SET name = ?, description = ?, latitude = ?, longitude = ? WHERE Id = ?`

const updatePoi = `UPDATE poi SET name = ?, description = ?, latitude = ?, longitude = ? WHERE Id = ?`

const getAllGuides = `SELECT Id,name, description, latitude, longitude FROM guide`

const getAllPois = `SELECT Id, name, description, latitude, longitude FROM poi WHERE guideid = ?`
