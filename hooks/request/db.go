package request

import (
	"encoding/base64"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/bandprotocol/chain/x/oracle/types"
)

func initDb(driverName, dataSourceName string) *gorm.DB {
	var db *gorm.DB
	var err error

	switch driverName {
	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		if err != nil {
			panic(fmt.Errorf("failed to connect to SQLite: %w", err))
		}
	case "postgres":
		db, err = gorm.Open(postgres.Open(dataSourceName), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		if err != nil {
			panic(fmt.Errorf("failed to connect to PostgreSQL: %w", err))
		}
	case "mysql":
		db, err = gorm.Open(mysql.Open(dataSourceName), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		if err != nil {
			panic(fmt.Errorf("failed to connect to MySQL: %w", err))
		}

	default:
		panic(fmt.Sprintf("unknown driver %s", driverName))
	}
	if err = db.AutoMigrate(&Request{}, &RequestedValidator{}, &RawReport{}, &RawRequest{}, &Report{}); err != nil {
		panic(fmt.Errorf("unable to auto-migrate DB: %w", err))
	}

	return db
}

func (h *Hook) insertRequests(requests []types.QueryRequestResponse) {
	var dbRequests []Request
	for _, request := range requests {
		dbRequests = append(dbRequests, GenerateRequestModel(request))
	}

	h.trans.Create(&dbRequests)
}

func (h *Hook) insertReports(reportMap map[types.RequestID][]types.Report) {
	var results []Report
	for requestID, reports := range reportMap {
		for _, report := range reports {
			results = append(results, GenerateReportModel(requestID, report))
		}
	}

	h.trans.Model(&Report{}).Create(&results)
}

func (h *Hook) removeOldRecords(request types.QueryRequestResponse) {
	if h.dbMaxRecords <= 0 {
		return
	}

	dbRequest := GenerateRequestModel(request)
	queryCondition := Request{
		OracleScriptID: dbRequest.OracleScriptID,
		Calldata:       dbRequest.Calldata,
		MinCount:       dbRequest.MinCount,
		AskCount:       dbRequest.AskCount,
	}

	// Keep the top `dbMaxRecords` records and delete the rest from database
	// under given search query
	subQuery := h.trans.
		Select("id").
		Where(&queryCondition).
		Order("id desc").
		Table("requests").
		Limit(h.dbMaxRecords)
	h.trans.
		Unscoped().
		Where(&queryCondition).
		Not("id IN (?)", subQuery).
		Delete(&Request{})
}

func (h *Hook) getLatestRequest(oid types.OracleScriptID, calldata []byte, askCount uint64, minCount uint64) (*Request, error) {
	var result Request
	queryCondition := Request{
		OracleScriptID: int64(oid),
		Calldata:       base64.StdEncoding.EncodeToString(calldata),
		AskCount:       askCount,
		MinCount:       minCount,
	}

	// Query latest request based on given search query
	queryResult := h.db.Model(&Request{}).
		Preload("Reports.RawReports").
		Preload(clause.Associations).
		Where(&queryCondition).
		Order("resolve_time desc").
		First(&result)

	if queryResult.Error != nil {
		return nil, fmt.Errorf("unable to query requests from searching database: %w", queryResult.Error)
	}

	return &result, nil
}
