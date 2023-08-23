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
	DeleteGuide(int64) error
	GetAllGuides() []guide
	Search(string) ([]guide, error)

	GetPoi(int64, int64) (*pointOfInterest, error)
	CreatePoi(*pointOfInterest) error
	UpdatePoi(*pointOfInterest) error
	DeletePoi(int64, int64) error
	GetAllPois(int64) []pointOfInterest
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

	//todo actually now that I think about it. Is this the migration part of a webapp?
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
	defer stmt.Close()

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
	defer stmt.Close()
	_, err = stmt.Exec(g.Name, g.Description, g.Coordinate.Latitude, g.Coordinate.Longitude, g.Id)
	if err != nil {
		return err
	}
	return nil
}

func (s *sqliteStore) DeleteGuide(id int64) error {
	stmt, err := s.db.Prepare(deleteGuide)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
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

func (s *sqliteStore) CreatePoi(poi *pointOfInterest) error {
	stmt, err := s.db.Prepare(insertPoi)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rs, err := stmt.Exec(poi.Name, poi.Description, poi.Coordinate.Latitude, poi.Coordinate.Longitude, poi.GuideID)
	if err != nil {
		return err
	}
	lastInsertID, err := rs.LastInsertId()
	if err != nil {
		return err
	}
	poi.Id = lastInsertID
	return nil
}

func (s *sqliteStore) UpdatePoi(poi *pointOfInterest) error {
	stmt, err := s.db.Prepare(updatePoi)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(poi.Name, poi.Description, poi.Coordinate.Latitude, poi.Coordinate.Longitude, poi.Id)
	if err != nil {
		return err
	}
	return nil
}

func (s *sqliteStore) GetPoi(guideID, poiID int64) (*pointOfInterest, error) {
	var (
		name        string
		description string
		latitude    float64
		longitude   float64
	)
	err := s.db.QueryRow(getPoi, guideID, poiID).Scan(&name, &description, &latitude, &longitude)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		p := pointOfInterest{
			Id:      poiID,
			GuideID: guideID,
			Coordinate: coordinate{
				Latitude:  latitude,
				Longitude: longitude,
			},
			Name:        name,
			Description: description,
		}
		return &p, nil

	}
}

func (s *sqliteStore) DeletePoi(guideId, poiID int64) error {
	stmt, err := s.db.Prepare(deletePoi)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(guideId, poiID)
	if err != nil {
		return err
	}
	return nil
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

func (s *sqliteStore) Search(query string) ([]guide, error) {
	rows, err := s.db.Query(searchGuides, query+"%")
	if err != nil {
		return nil, err
	}

	results := make([]guide, 0)
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
			return nil, err
		}
		g := guide{
			Id:          id,
			Name:        name,
			Description: description,
			Coordinate: coordinate{
				Latitude:  latitude,
				Longitude: longitude,
			},
		}
		results = append(results, g)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
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

const getPoi = `SELECT name, description, latitude, longitude FROM poi WHERE guideid = ? AND Id = ?`

const updateGuide = `UPDATE guide SET name = ?, description = ?, latitude = ?, longitude = ? WHERE Id = ?`

const updatePoi = `UPDATE poi SET name = ?, description = ?, latitude = ?, longitude = ? WHERE Id = ?`

const deleteGuide = `DELETE FROM guide WHERE Id = ?`

const deletePoi = `DELETE FROM poi WHERE guideid =? AND Id = ?`

const getAllGuides = `SELECT Id,name, description, latitude, longitude FROM guide`

const getAllPois = `SELECT Id, name, description, latitude, longitude FROM poi WHERE guideid = ?`

const searchGuides = `SELECT Id,name, description, latitude, longitude FROM guide WHERE name LIKE ?`
