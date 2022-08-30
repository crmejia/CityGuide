package guide

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
)

type store interface {
	Get(int64) (Guide, error) //TODO change return to (Guide,error) should allow the user to change the value directly, should use method
	Create(guide Guide) (int64, error)
	Update(guide Guide) error
	GetAllGuides() []Guide
}
type memoryStore struct {
	Guides  map[int64]Guide
	nextKey int64
}

func OpenMemoryStore() memoryStore {
	ms := memoryStore{
		Guides:  map[int64]Guide{},
		nextKey: 1,
	}
	return ms
}

func (s *memoryStore) Get(id int64) (Guide, error) {
	guide, ok := s.Guides[id]
	if ok {
		return guide, nil
	}
	return Guide{}, errors.New("guide not found")
}

func (s *memoryStore) Create(g Guide) (int64, error) {
	g.Id = s.nextKey
	s.Guides[g.Id] = g
	s.nextKey++
	return g.Id, nil
}

func (s *memoryStore) Update(g Guide) error {
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

	//_, err = db.Exec(createPoiTable)
	//if err != nil {
	//	return sqliteStore{}, err
	//}

	store := sqliteStore{
		db: db,
	}
	return store, nil
}

func (s *sqliteStore) Create(g Guide) (int64, error) {
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

func (s *sqliteStore) Get(id int64) (Guide, error) {
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

const createGuideTable = `
CREATE TABLE IF NOT EXISTS guide(
id INTEGER NOT NULL PRIMARY KEY,
name VARCHAR  NOT NULL,
description VARCHAR,
latitude float NOT NULL,
longitude float NOT NULL);`

const createPoiTable = `
CREATE TABLE IF NOT EXISTS poi(
name VARCHAR  NOT NULL,
description VARCHAR,
latitude float NOT NULL,
longitude float NOT NULL
guideId INTEGER NOT NULL,
FOREIGN_KEY(guideId) REFERENCES guide(guideId));`

const pragmaWALEnabled = `PRAGMA journal_mode = WAL;`
const pragma500BusyTimeout = `PRAGMA busy_timeout = 5000;`
const pragmaForeignKeysON = `PRAGMA foreign_keys = on=;`

const insertGuide = `INSERT INTO guide(name, description, latitude, longitude ) VALUES (?, ?, ?, ?);`

//const insertPoi = `INSERT INTO poi(name, description, latitude, longitude, guideId ) VALUES (?, ?, ?, ?, ?);`

const getGuide = `SELECT name, description, latitude, longitude FROM guide WHERE id = ?`

//const getPoi = `SELECT name, description, latitude, longitude FROM guide WHERE guideId = ?`

const updateGuide = `UPDATE guide SET name = ?, description = ?, latitude = ?, longitude = ? WHERE id = ?`
