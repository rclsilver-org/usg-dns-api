package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"

	"github.com/rclsilver-org/usg-dns-api/db"
)

func (s *Server) recordList(c *gin.Context) ([]db.Record, error) {
	return s.db.GetRecords(), nil
}

type recordGetIn struct {
	ID string `path:"record_id"`
}

func (s *Server) recordGet(c *gin.Context, in *recordGetIn) (*db.Record, error) {
	rec, err := s.db.GetRecord(in.ID)
	if err != nil {
		if err == db.ErrNotFound {
			return nil, errors.NewNotFound(nil, "no record found with this ID")
		}
		return nil, fmt.Errorf("error while fetching the record: %w", err)
	}

	return &rec, nil
}

type recordAddIn struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

func (s *Server) recordAdd(c *gin.Context, in *recordAddIn) (*db.Record, error) {
	rec, err := s.db.AddRecord(in.Name, in.Target)
	if err != nil {
		if err == db.ErrAlreadyExists {
			return nil, errors.NewAlreadyExists(err, "this record already exists")
		}
		return nil, fmt.Errorf("error while adding the record: %w", err)
	}

	return &rec, nil
}

type recordUpdateIn struct {
	ID     string `path:"record_id"`
	Name   string `json:"name"`
	Target string `json:"target"`
}

func (s *Server) recordUpdate(c *gin.Context, in *recordUpdateIn) (*db.Record, error) {
	rec, err := s.db.UpdateRecord(in.ID, in.Name, in.Target)
	if err != nil {
		if err == db.ErrNotFound {
			return nil, errors.NewNotFound(nil, "no record found with this ID")
		} else if err == db.ErrAlreadyExists {
			return nil, errors.NewAlreadyExists(nil, "a record already exists with those parameters")
		}
		return nil, fmt.Errorf("error while updating the record: %w", err)
	}

	return &rec, nil
}

type recordDeleteIn struct {
	ID string `path:"record_id"`
}

func (s *Server) recordDelete(c *gin.Context, in *recordDeleteIn) error {

	if err := s.db.DeleteRecord(in.ID); err != nil {
		if err == db.ErrNotFound {
			return errors.NewNotFound(nil, "no record found with this ID")
		}
		return fmt.Errorf("error while deleting the record: %w", err)
	}

	return nil
}
