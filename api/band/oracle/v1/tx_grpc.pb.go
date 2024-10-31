// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: band/oracle/v1/tx.proto

package oraclev1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Msg_RequestData_FullMethodName        = "/band.oracle.v1.Msg/RequestData"
	Msg_ReportData_FullMethodName         = "/band.oracle.v1.Msg/ReportData"
	Msg_CreateDataSource_FullMethodName   = "/band.oracle.v1.Msg/CreateDataSource"
	Msg_EditDataSource_FullMethodName     = "/band.oracle.v1.Msg/EditDataSource"
	Msg_CreateOracleScript_FullMethodName = "/band.oracle.v1.Msg/CreateOracleScript"
	Msg_EditOracleScript_FullMethodName   = "/band.oracle.v1.Msg/EditOracleScript"
	Msg_Activate_FullMethodName           = "/band.oracle.v1.Msg/Activate"
	Msg_UpdateParams_FullMethodName       = "/band.oracle.v1.Msg/UpdateParams"
)

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MsgClient interface {
	// RequestData defines a method for submitting a new request.
	RequestData(ctx context.Context, in *MsgRequestData, opts ...grpc.CallOption) (*MsgRequestDataResponse, error)
	// ReportData defines a method for reporting a data to resolve the request.
	ReportData(ctx context.Context, in *MsgReportData, opts ...grpc.CallOption) (*MsgReportDataResponse, error)
	// CreateDataSource defines a method for creating a new data source.
	CreateDataSource(ctx context.Context, in *MsgCreateDataSource, opts ...grpc.CallOption) (*MsgCreateDataSourceResponse, error)
	// EditDataSource defines a method for editing an existing data source.
	EditDataSource(ctx context.Context, in *MsgEditDataSource, opts ...grpc.CallOption) (*MsgEditDataSourceResponse, error)
	// CreateOracleScript defines a method for creating a new oracle script.
	CreateOracleScript(ctx context.Context, in *MsgCreateOracleScript, opts ...grpc.CallOption) (*MsgCreateOracleScriptResponse, error)
	// EditOracleScript defines a method for editing an existing oracle script.
	EditOracleScript(ctx context.Context, in *MsgEditOracleScript, opts ...grpc.CallOption) (*MsgEditOracleScriptResponse, error)
	// Activate defines a method for applying to be an oracle validator.
	Activate(ctx context.Context, in *MsgActivate, opts ...grpc.CallOption) (*MsgActivateResponse, error)
	// UpdateParams defines a governance operation for updating the x/oracle module
	// parameters.
	//
	// Since: cosmos-sdk 0.47
	UpdateParams(ctx context.Context, in *MsgUpdateParams, opts ...grpc.CallOption) (*MsgUpdateParamsResponse, error)
}

type msgClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgClient(cc grpc.ClientConnInterface) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) RequestData(ctx context.Context, in *MsgRequestData, opts ...grpc.CallOption) (*MsgRequestDataResponse, error) {
	out := new(MsgRequestDataResponse)
	err := c.cc.Invoke(ctx, Msg_RequestData_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ReportData(ctx context.Context, in *MsgReportData, opts ...grpc.CallOption) (*MsgReportDataResponse, error) {
	out := new(MsgReportDataResponse)
	err := c.cc.Invoke(ctx, Msg_ReportData_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) CreateDataSource(ctx context.Context, in *MsgCreateDataSource, opts ...grpc.CallOption) (*MsgCreateDataSourceResponse, error) {
	out := new(MsgCreateDataSourceResponse)
	err := c.cc.Invoke(ctx, Msg_CreateDataSource_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) EditDataSource(ctx context.Context, in *MsgEditDataSource, opts ...grpc.CallOption) (*MsgEditDataSourceResponse, error) {
	out := new(MsgEditDataSourceResponse)
	err := c.cc.Invoke(ctx, Msg_EditDataSource_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) CreateOracleScript(ctx context.Context, in *MsgCreateOracleScript, opts ...grpc.CallOption) (*MsgCreateOracleScriptResponse, error) {
	out := new(MsgCreateOracleScriptResponse)
	err := c.cc.Invoke(ctx, Msg_CreateOracleScript_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) EditOracleScript(ctx context.Context, in *MsgEditOracleScript, opts ...grpc.CallOption) (*MsgEditOracleScriptResponse, error) {
	out := new(MsgEditOracleScriptResponse)
	err := c.cc.Invoke(ctx, Msg_EditOracleScript_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) Activate(ctx context.Context, in *MsgActivate, opts ...grpc.CallOption) (*MsgActivateResponse, error) {
	out := new(MsgActivateResponse)
	err := c.cc.Invoke(ctx, Msg_Activate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateParams(ctx context.Context, in *MsgUpdateParams, opts ...grpc.CallOption) (*MsgUpdateParamsResponse, error) {
	out := new(MsgUpdateParamsResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateParams_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
// All implementations must embed UnimplementedMsgServer
// for forward compatibility
type MsgServer interface {
	// RequestData defines a method for submitting a new request.
	RequestData(context.Context, *MsgRequestData) (*MsgRequestDataResponse, error)
	// ReportData defines a method for reporting a data to resolve the request.
	ReportData(context.Context, *MsgReportData) (*MsgReportDataResponse, error)
	// CreateDataSource defines a method for creating a new data source.
	CreateDataSource(context.Context, *MsgCreateDataSource) (*MsgCreateDataSourceResponse, error)
	// EditDataSource defines a method for editing an existing data source.
	EditDataSource(context.Context, *MsgEditDataSource) (*MsgEditDataSourceResponse, error)
	// CreateOracleScript defines a method for creating a new oracle script.
	CreateOracleScript(context.Context, *MsgCreateOracleScript) (*MsgCreateOracleScriptResponse, error)
	// EditOracleScript defines a method for editing an existing oracle script.
	EditOracleScript(context.Context, *MsgEditOracleScript) (*MsgEditOracleScriptResponse, error)
	// Activate defines a method for applying to be an oracle validator.
	Activate(context.Context, *MsgActivate) (*MsgActivateResponse, error)
	// UpdateParams defines a governance operation for updating the x/oracle module
	// parameters.
	//
	// Since: cosmos-sdk 0.47
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) RequestData(context.Context, *MsgRequestData) (*MsgRequestDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestData not implemented")
}
func (UnimplementedMsgServer) ReportData(context.Context, *MsgReportData) (*MsgReportDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportData not implemented")
}
func (UnimplementedMsgServer) CreateDataSource(context.Context, *MsgCreateDataSource) (*MsgCreateDataSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateDataSource not implemented")
}
func (UnimplementedMsgServer) EditDataSource(context.Context, *MsgEditDataSource) (*MsgEditDataSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EditDataSource not implemented")
}
func (UnimplementedMsgServer) CreateOracleScript(context.Context, *MsgCreateOracleScript) (*MsgCreateOracleScriptResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateOracleScript not implemented")
}
func (UnimplementedMsgServer) EditOracleScript(context.Context, *MsgEditOracleScript) (*MsgEditOracleScriptResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EditOracleScript not implemented")
}
func (UnimplementedMsgServer) Activate(context.Context, *MsgActivate) (*MsgActivateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Activate not implemented")
}
func (UnimplementedMsgServer) UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateParams not implemented")
}
func (UnimplementedMsgServer) mustEmbedUnimplementedMsgServer() {}

// UnsafeMsgServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MsgServer will
// result in compilation errors.
type UnsafeMsgServer interface {
	mustEmbedUnimplementedMsgServer()
}

func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	s.RegisterService(&Msg_ServiceDesc, srv)
}

func _Msg_RequestData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRequestData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RequestData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_RequestData_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RequestData(ctx, req.(*MsgRequestData))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ReportData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgReportData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ReportData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_ReportData_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ReportData(ctx, req.(*MsgReportData))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CreateDataSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateDataSource)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateDataSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_CreateDataSource_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateDataSource(ctx, req.(*MsgCreateDataSource))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_EditDataSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgEditDataSource)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).EditDataSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_EditDataSource_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).EditDataSource(ctx, req.(*MsgEditDataSource))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CreateOracleScript_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateOracleScript)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateOracleScript(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_CreateOracleScript_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateOracleScript(ctx, req.(*MsgCreateOracleScript))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_EditOracleScript_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgEditOracleScript)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).EditOracleScript(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_EditOracleScript_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).EditOracleScript(ctx, req.(*MsgEditOracleScript))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_Activate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgActivate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Activate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_Activate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Activate(ctx, req.(*MsgActivate))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateParams_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateParams(ctx, req.(*MsgUpdateParams))
	}
	return interceptor(ctx, in, info, handler)
}

// Msg_ServiceDesc is the grpc.ServiceDesc for Msg service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Msg_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "band.oracle.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequestData",
			Handler:    _Msg_RequestData_Handler,
		},
		{
			MethodName: "ReportData",
			Handler:    _Msg_ReportData_Handler,
		},
		{
			MethodName: "CreateDataSource",
			Handler:    _Msg_CreateDataSource_Handler,
		},
		{
			MethodName: "EditDataSource",
			Handler:    _Msg_EditDataSource_Handler,
		},
		{
			MethodName: "CreateOracleScript",
			Handler:    _Msg_CreateOracleScript_Handler,
		},
		{
			MethodName: "EditOracleScript",
			Handler:    _Msg_EditOracleScript_Handler,
		},
		{
			MethodName: "Activate",
			Handler:    _Msg_Activate_Handler,
		},
		{
			MethodName: "UpdateParams",
			Handler:    _Msg_UpdateParams_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "band/oracle/v1/tx.proto",
}