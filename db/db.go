package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"

	"github.com/rclsilver-org/usg-dns-api/pkg/utils"
)

var (
	ErrAlreadyExists = errors.New("resource already exists")
	ErrNotFound      = errors.New("resource not found")
)

type Database struct {
	cfg *config
	mut sync.Mutex

	data struct {
		MasterToken string `json:"master-token"`

		Records []Record `json:"records"`
	}
}

func NewDatabase(ctx context.Context) (*Database, error) {
	// load the configuration
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to load the configuration: %w", err)
	}
	logrus.WithContext(ctx).Debug("loaded the database configuration")

	db := &Database{}
	db.data.Records = make([]Record, 0)

	f, err := os.Open(cfg.Path)
	if err == nil {
		defer f.Close()

		data, err := io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("unable to read the database: %w", err)
		}

		if err := json.Unmarshal(data, &db.data); err != nil {
			return nil, fmt.Errorf("unable to unmarshal the data: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("unable to load the database: %w", err)
	}

	db.cfg = cfg

	return db, nil
}

func (db *Database) GenerateMasterToken() string {
	db.mut.Lock()
	defer db.mut.Unlock()

	token := uuid.NewString()
	hash := utils.StringHash(token)

	db.data.MasterToken = hash

	return token
}

func (db *Database) GetMasterToken() string {
	db.mut.Lock()
	defer db.mut.Unlock()

	return db.data.MasterToken
}

func (db *Database) GetRecord(id string) (Record, error) {
	if err := validateID(id); err != nil {
		return Record{}, err
	}

	db.mut.Lock()
	defer db.mut.Unlock()

	for _, record := range db.data.Records {
		if record.ID == id {
			return record, nil
		}
	}

	return Record{}, ErrNotFound
}

func (db *Database) AddRecord(name, target string) (Record, error) {
	if err := validateName(name); err != nil {
		return Record{}, err
	}

	if err := validateTarget(target); err != nil {
		return Record{}, err
	}

	db.mut.Lock()
	defer db.mut.Unlock()

	r := Record{
		Name:   name,
		Target: target,
	}

	for {
		r.ID = uuid.NewString()
		found := false

		for _, record := range db.data.Records {
			if record.Name == r.Name {
				return Record{}, ErrAlreadyExists
			}

			if record.ID == r.ID {
				found = true
				break
			}
		}

		if !found {
			break
		}
	}

	db.data.Records = append(db.data.Records, r)

	if err := db.save(); err != nil {
		return Record{}, err
	}

	return r, nil
}

func (db *Database) UpdateRecord(id, name, target string) (Record, error) {
	if err := validateID(id); err != nil {
		return Record{}, err
	}

	if err := validateName(name); err != nil {
		return Record{}, err
	}

	if err := validateTarget(target); err != nil {
		return Record{}, err
	}

	db.mut.Lock()
	defer db.mut.Unlock()

	for i, record := range db.data.Records {
		if record.ID == id {
			for _, rec := range db.data.Records {
				if rec.Name == name && rec.ID != id {
					return Record{}, ErrAlreadyExists
				}
			}

			db.data.Records[i].Name = name
			db.data.Records[i].Target = target

			if err := db.save(); err != nil {
				return Record{}, err
			}

			return db.data.Records[i], nil
		}
	}

	return Record{}, ErrNotFound
}

func (db *Database) DeleteRecord(id string) error {
	if err := validateID(id); err != nil {
		return err
	}

	db.mut.Lock()
	defer db.mut.Unlock()

	for i, record := range db.data.Records {
		if record.ID == id {
			db.data.Records = append(db.data.Records[:i], db.data.Records[i+1:]...)

			if err := db.save(); err != nil {
				return err
			}

			return nil
		}
	}

	return ErrNotFound
}

func (db *Database) GetRecords() []Record {
	db.mut.Lock()
	defer db.mut.Unlock()

	recordsCopy := make([]Record, len(db.data.Records))
	copy(recordsCopy, db.data.Records)
	return recordsCopy
}

func (db *Database) save() error {
	data, err := json.MarshalIndent(db.data, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal the data: %w", err)
	}

	f, err := os.Create(db.cfg.Path)
	if err != nil {
		return fmt.Errorf("unable to create the database file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("unable to write the database file: %w", err)
	}

	return nil
}

func (db *Database) Save() error {
	db.mut.Lock()
	defer db.mut.Unlock()

	return db.save()
}
