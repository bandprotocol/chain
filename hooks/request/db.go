package request

import (
	"encoding/base64"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/bandprotocol/chain/x/oracle/types"
)

type RawReport struct {
	gorm.Model
	ReportID   uint
	ExternalID int64
	Data       string `gorm:"size:1024"`
	ExitCode   uint32
}

type Report struct {
	gorm.Model
	RequestID       uint
	Validator       string
	RawReports      []RawReport
	InBeforeResolve bool
}

type RawRequest struct {
	gorm.Model
	RequestID    uint
	ExternalID   int64
	DataSourceID int64
	CallData     string `gorm:"size:1024"`
}

type RequestedValidator struct {
	gorm.Model
	RequestID uint
	Address   string
}

type Request struct {
	gorm.Model
	OracleScriptID      int64  `gorm:"index:idx_min_count_ask_count_oracle_script_id_calldata"`
	CallData            string `gorm:"index:idx_min_count_ask_count_oracle_script_id_calldata;size:1024"`
	RequestedValidators []RequestedValidator
	MinCount            uint64 `gorm:"index:idx_min_count_ask_count_oracle_script_id_calldata"`
	AskCount            uint64 `gorm:"index:idx_min_count_ask_count_oracle_script_id_calldata"`
	AnsCount            uint64
	RequestHeight       int64
	RequestTime         time.Time
	ClientID            string
	RawRequests         []RawRequest
	IBCPortID           string
	IBCChannelID        string
	ExecuteGas          uint64
	Reports             []Report
	ResolveTime         time.Time `gorm:"index:idx_min_count_ask_count_oracle_script_id_calldata"`
	ResolveStatus       string    `gorm:"size:64"`
	ResultCallData      string    `gorm:"size:1024"`
	Result              string    `gorm:"size:1024"`
}

func (r Request) QueryRequestResponse() types.QueryRequestResponse {
	// Request's calldata
	callData, err := base64.StdEncoding.DecodeString(r.CallData)
	if err != nil {
		panic(err)
	}

	// Requested validators
	var requestedValidators []string
	for _, rVal := range r.RequestedValidators {
		requestedValidators = append(requestedValidators, rVal.Address)
	}

	// Raw requests
	var rawRequests []types.RawRequest
	for _, rr := range r.RawRequests {
		callData, err := base64.StdEncoding.DecodeString(r.CallData)
		if err != nil {
			panic(err)
		}
		rawRequests = append(rawRequests, types.RawRequest{
			ExternalID:   types.ExternalID(rr.ExternalID),
			DataSourceID: types.DataSourceID(rr.DataSourceID),
			Calldata:     callData,
		})
	}

	// Result's calldata
	resultCallData, err := base64.StdEncoding.DecodeString(r.ResultCallData)
	if err != nil {
		panic(err)
	}

	// Result data
	requestResult, err := base64.StdEncoding.DecodeString(r.Result)
	if err != nil {
		panic(err)
	}

	// Reports
	var reports []types.Report
	for _, dbReport := range r.Reports {
		// Raw reports
		var rawReports []types.RawReport
		for _, rr := range dbReport.RawReports {
			data, err := base64.StdEncoding.DecodeString(rr.Data)
			if err != nil {
				panic(err)
			}
			rawReports = append(rawReports, types.RawReport{
				ExternalID: types.ExternalID(rr.ExternalID),
				Data:       data,
				ExitCode:   rr.ExitCode,
			})
		}
		report := types.Report{
			Validator:       dbReport.Validator,
			InBeforeResolve: dbReport.InBeforeResolve,
			RawReports:      rawReports,
		}
		reports = append(reports, report)
	}

	// IBC Channel
	var ibcChannel *types.IBCChannel
	if len(r.IBCChannelID) > 0 {
		ibcChannel = &types.IBCChannel{
			PortId:    r.IBCPortID,
			ChannelId: r.IBCChannelID,
		}
	}

	// The whole response
	request := types.QueryRequestResponse{
		Request: &types.Request{
			OracleScriptID:      types.OracleScriptID(r.OracleScriptID),
			Calldata:            callData,
			MinCount:            r.MinCount,
			RequestHeight:       r.RequestHeight,
			RequestTime:         uint64(r.RequestTime.Unix()),
			ClientID:            r.ClientID,
			IBCChannel:          ibcChannel,
			ExecuteGas:          r.ExecuteGas,
			RequestedValidators: requestedValidators,
			RawRequests:         rawRequests,
		},
		Result: &types.Result{
			ClientID:       r.ClientID,
			OracleScriptID: types.OracleScriptID(r.OracleScriptID),
			Calldata:       resultCallData,
			AskCount:       r.AskCount,
			MinCount:       r.MinCount,
			AnsCount:       r.AnsCount,
			RequestID:      types.RequestID(r.ID),
			RequestTime:    r.RequestTime.Unix(),
			ResolveTime:    r.ResolveTime.Unix(),
			ResolveStatus:  types.ResolveStatus(types.ResolveStatus_value[r.ResolveStatus]),
			Result:         requestResult,
		},
		Reports: reports,
	}

	return request
}

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

func generateRequestModel(data types.QueryRequestResponse) Request {
	request := data.Request
	reports := data.Reports
	result := data.Result

	dbRequest := Request{
		Model: gorm.Model{
			ID: uint(result.RequestID),
		},
		OracleScriptID: int64(request.OracleScriptID),
		CallData:       base64.StdEncoding.EncodeToString(request.Calldata),
		MinCount:       result.MinCount,
		AskCount:       result.AskCount,
		AnsCount:       result.AnsCount,
		RequestHeight:  request.RequestHeight,
		RequestTime:    time.Unix(int64(request.RequestTime), 0),
		ClientID:       result.ClientID,
		ResolveTime:    time.Unix(result.ResolveTime, 0),
		ResolveStatus:  result.ResolveStatus.String(),
		ExecuteGas:     request.ExecuteGas,
		ResultCallData: base64.StdEncoding.EncodeToString(result.Calldata),
		Result:         base64.StdEncoding.EncodeToString(result.Result),
	}

	if request.IBCChannel != nil {
		dbRequest.IBCChannelID = request.IBCChannel.ChannelId
		dbRequest.IBCPortID = request.IBCChannel.PortId
	}

	for _, reqVal := range request.RequestedValidators {
		dbRequest.RequestedValidators = append(dbRequest.RequestedValidators, RequestedValidator{
			Address: reqVal,
		})
	}

	for _, rawReq := range request.RawRequests {
		dbRequest.RawRequests = append(dbRequest.RawRequests, RawRequest{
			ExternalID:   int64(rawReq.ExternalID),
			DataSourceID: int64(rawReq.DataSourceID),
			CallData:     base64.StdEncoding.EncodeToString(rawReq.Calldata),
		})
	}

	for _, report := range reports {
		var rawReports []RawReport
		for _, rawReport := range report.RawReports {
			rawReports = append(rawReports, RawReport{
				ExternalID: int64(rawReport.ExternalID),
				Data:       base64.StdEncoding.EncodeToString(rawReport.Data),
				ExitCode:   rawReport.ExitCode,
			})
		}
		dbRequest.Reports = append(dbRequest.Reports, Report{
			Validator:       report.Validator,
			RawReports:      rawReports,
			InBeforeResolve: report.InBeforeResolve,
		})
	}

	return dbRequest
}

func generateReportModel(requestID types.RequestID, report types.Report) Report {
	result := Report{
		RequestID:       uint(requestID),
		Validator:       report.Validator,
		InBeforeResolve: report.InBeforeResolve,
	}

	for _, rawReport := range report.RawReports {
		result.RawReports = append(result.RawReports, RawReport{
			ExternalID: int64(rawReport.ExternalID),
			Data:       base64.StdEncoding.EncodeToString(rawReport.Data),
			ExitCode:   rawReport.ExitCode,
		})
	}

	return result
}

func (h *Hook) insertRequests(requests []types.QueryRequestResponse) {
	var dbRequests []Request
	for _, request := range requests {
		dbRequests = append(dbRequests, generateRequestModel(request))
	}

	h.trans.Create(&dbRequests)
}

func (h *Hook) insertReports(reportMap map[types.RequestID][]types.Report) {
	var results []Report

	for requestID, reports := range reportMap {
		for _, report := range reports {
			results = append(results, generateReportModel(requestID, report))
		}
		h.trans.
			Model(&Request{
				Model: gorm.Model{
					ID: uint(requestID),
				},
			}).
			Update("ans_count", gorm.Expr("ans_count + ?", len(reports)))
	}

	h.trans.Model(&Report{}).Create(&results)
}

func (h *Hook) removeOldRecords() {
	if h.dbMaxRecords <= 0 {
		return
	}
	subQuery := h.trans.Select("id").Order("id desc").Table("requests").Limit(h.dbMaxRecords)
	h.trans.Unscoped().Not("id IN (?)", subQuery).Delete(&Request{})
}

func (h *Hook) getLatestRequest(oid types.OracleScriptID, calldata []byte, askCount uint64, minCount uint64) (*Request, error) {
	queryCondition := Request{
		OracleScriptID: int64(oid),
		CallData:       base64.StdEncoding.EncodeToString(calldata),
		AskCount:       askCount,
		MinCount:       minCount,
	}
	var result Request
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
