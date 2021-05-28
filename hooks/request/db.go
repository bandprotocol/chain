package request

import (
	"encoding/base64"
	"fmt"
	"time"

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

type Requests []Request

func (r Request) QueryRequestResponse() types.QueryRequestResponse {
	callData, err := base64.StdEncoding.DecodeString(r.CallData)
	if err != nil {
		panic(err)
	}

	var requestedValidators []string
	for _, rVal := range r.RequestedValidators {
		requestedValidators = append(requestedValidators, rVal.Address)
	}
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
	resultCallData, err := base64.StdEncoding.DecodeString(r.ResultCallData)
	if err != nil {
		panic(err)
	}
	requestResult, err := base64.StdEncoding.DecodeString(r.Result)
	if err != nil {
		panic(err)
	}

	var reports []types.Report
	for _, dbReport := range r.Reports {
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

	var ibcChannel *types.IBCChannel
	if len(r.IBCChannelID) > 0 {
		ibcChannel = &types.IBCChannel{
			PortId:    r.IBCPortID,
			ChannelId: r.IBCChannelID,
		}
	}

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

func (r Requests) QueryRequestSearchResponse() types.QueryRequestSearchResponse {
	var finalResult types.QueryRequestSearchResponse
	for _, dbReq := range r {
		request := dbReq.QueryRequestResponse()
		finalResult.Requests = append(finalResult.Requests, &request)
	}

	return finalResult
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
	default:
		panic(fmt.Sprintf("unknown driver %s", driverName))
	}
	if err = db.AutoMigrate(&Request{}, &RequestedValidator{}, &RawReport{}, &RawRequest{}, &Report{}); err != nil {
		panic(fmt.Errorf("unable to auto-migrate DB: %w", err))
	}

	return db
}

func (h *Hook) insertRequest(request types.Request, reports []types.Report, result types.Result) {
	dbData := Request{
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
		dbData.IBCChannelID = request.IBCChannel.ChannelId
		dbData.IBCPortID = request.IBCChannel.PortId
	}

	for _, reqVal := range request.RequestedValidators {
		dbData.RequestedValidators = append(dbData.RequestedValidators, RequestedValidator{
			Address: reqVal,
		})
	}

	for _, rawReq := range request.RawRequests {
		dbData.RawRequests = append(dbData.RawRequests, RawRequest{
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
		dbData.Reports = append(dbData.Reports, Report{
			Validator:       report.Validator,
			RawReports:      rawReports,
			InBeforeResolve: report.InBeforeResolve,
		})
	}

	h.trans.Create(&dbData)
}

func (h *Hook) addReport(requestIDs types.RequestID, report types.Report) {
	result := Report{
		RequestID:       uint(requestIDs),
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
	h.trans.Model(&Report{}).Create(&result)
}

func (h *Hook) getMultiRequests(oid types.OracleScriptID, calldata []byte, askCount uint64, minCount uint64, limit uint64) (Requests, error) {
	queryCondition := Request{
		OracleScriptID: int64(oid),
		CallData:       base64.StdEncoding.EncodeToString(calldata),
		AskCount:       askCount,
		MinCount:       minCount,
	}
	var result []Request
	queryResult := h.db.Model(&Request{}).Limit(int(limit)).Preload("Reports.RawReports").Preload(clause.Associations).Where(&queryCondition).Order("resolve_time desc").Find(&result)
	if queryResult.Error != nil {
		return nil, fmt.Errorf("unable to query requests from searching database: %w", queryResult.Error)
	}
	return result, nil
}
