package repo

import (
	"fmt"
	"scratch/core/domain"
	"scratch/core/port"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/gin-gonic/gin"
)

type ScratchRepo struct {
	db *DBStruct
}

func NewScratchRepo(db *DBStruct) *ScratchRepo {
	return &ScratchRepo{
		db,
	}
}

func (sr *ScratchRepo) CreateScratch(c *gin.Context, req *domain.Scratch) error {
	// Build SQL - use sq.Expr for NOW()
	query := psql.Insert("scratch").
		Columns("name", "lastupdated", "password").
		Values(req.Name, sq.Expr("NOW()"), req.Password).
		Suffix("RETURNING name, password") // <-- must match Scan below

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	fmt.Println("SQL:", sql)
	fmt.Println("Args:", args)

	// Execute and scan exactly the number of returned columns
	if err := sr.db.QueryRow(c, sql, args...).Scan(&req.Name, &req.Password); err != nil {
		fmt.Println("DB Error:", err) // helpful while debugging
		return err
	}
	return nil
}

func (sr *ScratchRepo) FetchScratch(c *gin.Context, name string) (*domain.Scratch, error) {
	// Build SQL - use sq.Expr for NOW()

	var rsp domain.Scratch
	query := psql.Select("*").From("scratch").Where(sq.Eq{"name": name})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	fmt.Println("SQL:", sql)
	fmt.Println("Args:", args)

	err = sr.db.QueryRow(c, sql, args...).Scan(
		&rsp.Id,
		&rsp.Name,
		&rsp.LastUpdatedDate,
		&rsp.Password,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, port.ErrDataNotFound
		}
		return nil, err
	}

	return &rsp, nil
}
