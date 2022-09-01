package guide

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
)

type store interface {
	GetGuide(int64) (guide, error)
	CreateGuide(string, ...guideOption) (*guide, error)
	UpdateGuide(*guide) error

	GetPoi(int64) (pointOfInterest, error)
	CreatePoi(string, int64, ...poiOption) (*pointOfInterest, error)
	UpdatePoi(*pointOfInterest) error

	GetAllGuides() []guide
	GetAllPois(int64) []pointOfInterest
}
type memoryStore struct {
	Guides       map[int64]guide
	Pois         map[int64]pointOfInterest
	NextGuideKey int64
	NextPoiKey   int64
}

func OpenMemoryStore() memoryStore {
	ms := memoryStore{
		Guides:       map[int64]guide{},
		Pois:         map[int64]pointOfInterest{},
		NextGuideKey: 1,
		NextPoiKey:   1,
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
	g, err := newGuide(name, opts...)
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
		return errors.New("must set the id of the guide")
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
	pois := []pointOfInterest{}
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

func OpenSQLiteStore(dbPath string) (sqliteStore, error) {
	if dbPath == "" {
		return sqliteStore{}, errors.New("db source cannot be empty")
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return sqliteStore{}, err
	}

	for _, stmt := range []string{pragmaWALEnabled, pragma500BusyTimeout, pragmaForeignKeysON} {
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

func (s *sqliteStore) CreateGuide(name string, opts ...guideOption) (*guide, error) {
	g, err := newGuide(name, opts...)
	if err != nil {
		return nil, err
	}
	stmt, err := s.db.Prepare(insertGuide)
	if err != nil {
		return nil, err
	}

	rs, err := stmt.Exec(g.Name, g.Description, g.Coordinate.Latitude, g.Coordinate.Longitude)
	if err != nil {
		return nil, err
	}
	lastInsertID, err := rs.LastInsertId()
	if err != nil {
		return nil, err
	}
	g.Id = lastInsertID
	return &g, nil
}

func (s *sqliteStore) GetGuide(id int64) (guide, error) {
	rows, err := s.db.Query(getGuide, id)
	if err != nil {
		return guide{}, err
	}

	g := guide{}

	for rows.Next() {
		var (
			name        string
			description string
			latitude    float64
			longitude   float64
		)
		err = rows.Scan(&name, &description, &latitude, &longitude)
		if err != nil {
			return guide{}, err
		}
		g.Id = id
		g.Name = name
		g.Description = description
		g.Coordinate = coordinate{Latitude: latitude, Longitude: longitude}
	}

	if err = rows.Err(); err != nil {
		return guide{}, err
	}

	return g, nil
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
id INTEGER NOT NULL PRIMARY KEY,
name TEXT  NOT NULL,
description TEXT,
latitude REAL NOT NULL,
longitude REAL NOT NULL,
CHECK (name <> ''));`

const createPoiTable = `
CREATE TABLE IF NOT EXISTS poi(
id INTEGER NOT NULL PRIMARY KEY,
name TEXT  NOT NULL,
description TEXT,
latitude REAL NOT NULL,
longitude REAL NOT NULL,
guideId INTEGER NOT NULL,
FOREIGN KEY(guideId) REFERENCES guide(id),
CHECK (name <> ''));`

const insertGuide = `INSERT INTO guide(name, description, latitude, longitude ) VALUES (?, ?, ?, ?);`

const insertPoi = `INSERT INTO poi(name, description, latitude, longitude, guideId ) VALUES (?, ?, ?, ?, ?);`

const getGuide = `SELECT name, description, latitude, longitude FROM guide WHERE id = ?`

const getPoi = `SELECT name, description, latitude, longitude, guideid FROM poi WHERE Id = ?`

const updateGuide = `UPDATE guide SET name = ?, description = ?, latitude = ?, longitude = ? WHERE id = ?`

const updatePoi = `UPDATE poi SET name = ?, description = ?, latitude = ?, longitude = ? WHERE id = ?`

const getAllGuides = `SELECT id,name, description, latitude, longitude FROM guide`

const getAllPois = `SELECT id, name, description, latitude, longitude FROM poi WHERE guideid = ?`
