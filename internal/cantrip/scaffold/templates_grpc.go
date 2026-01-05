package scaffold

import "fmt"

func grpcServerTemplate(name, namePascal, goModulePath string) string {
	return fmt.Sprintf(`package %s

import (
	"context"

	"github.com/google/uuid"

	pb "%s/gen/proto/v1"
)

type GRPCServer struct {
	pb.Unimplemented%sServiceServer
	service *Service
}

func NewGRPCServer(service *Service) *GRPCServer {
	return &GRPCServer{service: service}
}

func (s *GRPCServer) Create(ctx context.Context, req *pb.Create%sRequest) (*pb.Create%sResponse, error) {
	entity := &%s{
		Name: req.Name,
	}
	if err := s.service.Create(ctx, entity); err != nil {
		return nil, err
	}
	return &pb.Create%sResponse{
		%s: entityToProto(entity),
	}, nil
}

func (s *GRPCServer) Get(ctx context.Context, req *pb.Get%sRequest) (*pb.Get%sResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}
	entity, err := s.service.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &pb.Get%sResponse{
		%s: entityToProto(entity),
	}, nil
}

func (s *GRPCServer) List(ctx context.Context, req *pb.List%sRequest) (*pb.List%sResponse, error) {
	entities, err := s.service.List(ctx)
	if err != nil {
		return nil, err
	}
	var items []*pb.%s
	for _, e := range entities {
		items = append(items, entityToProto(e))
	}
	return &pb.List%sResponse{
		%ss: items,
	}, nil
}

func (s *GRPCServer) Update(ctx context.Context, req *pb.Update%sRequest) (*pb.Update%sResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}
	entity := &%s{
		ID:   id,
		Name: req.Name,
	}
	if err := s.service.Update(ctx, entity); err != nil {
		return nil, err
	}
	return &pb.Update%sResponse{
		%s: entityToProto(entity),
	}, nil
}

func (s *GRPCServer) Delete(ctx context.Context, req *pb.Delete%sRequest) (*pb.Delete%sResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}
	if err := s.service.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &pb.Delete%sResponse{}, nil
}

func entityToProto(e *%s) *pb.%s {
	return &pb.%s{
		Id:   e.ID.String(),
		Name: e.Name,
	}
}
`, name, goModulePath, namePascal,
		namePascal, namePascal, namePascal, namePascal, namePascal,
		namePascal, namePascal, namePascal, namePascal,
		namePascal, namePascal, namePascal, namePascal, namePascal,
		namePascal, namePascal, namePascal, namePascal, namePascal,
		namePascal, namePascal, namePascal,
		namePascal, namePascal, namePascal)
}

func protoTemplate(name, namePascal, goModulePath string) string {
	return fmt.Sprintf(`syntax = "proto3";

package %s.v1;

option go_package = "%s/gen/proto/v1;%sv1";

service %sService {
  rpc Create(Create%sRequest) returns (Create%sResponse);
  rpc Get(Get%sRequest) returns (Get%sResponse);
  rpc List(List%sRequest) returns (List%sResponse);
  rpc Update(Update%sRequest) returns (Update%sResponse);
  rpc Delete(Delete%sRequest) returns (Delete%sResponse);
}

message %s {
  string id = 1;
  string name = 2;
}

message Create%sRequest {
  string name = 1;
}

message Create%sResponse {
  %s %s = 1;
}

message Get%sRequest {
  string id = 1;
}

message Get%sResponse {
  %s %s = 1;
}

message List%sRequest {
  int32 limit = 1;
  int32 offset = 2;
}

message List%sResponse {
  repeated %s %ss = 1;
}

message Update%sRequest {
  string id = 1;
  string name = 2;
}

message Update%sResponse {
  %s %s = 1;
}

message Delete%sRequest {
  string id = 1;
}

message Delete%sResponse {}
`, name, goModulePath, name, namePascal,
		namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal,
		namePascal,
		namePascal, namePascal, namePascal, name,
		namePascal, namePascal, namePascal, name,
		namePascal, namePascal, namePascal, name,
		namePascal, namePascal, namePascal, name,
		namePascal, namePascal)
}
