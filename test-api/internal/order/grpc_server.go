package order

import (
	"context"

	"github.com/google/uuid"

	pb "github.com/test/api/gen/proto/v1"
)

type GRPCServer struct {
	pb.UnimplementedOrderServiceServer
	service *Service
}

func NewGRPCServer(service *Service) *GRPCServer {
	return &GRPCServer{service: service}
}

func (s *GRPCServer) Create(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	entity := &Order{
		Name: req.Name,
	}
	if err := s.service.Create(ctx, entity); err != nil {
		return nil, err
	}
	return &pb.CreateOrderResponse{
		Order: entityToProto(entity),
	}, nil
}

func (s *GRPCServer) Get(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}
	entity, err := s.service.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &pb.GetOrderResponse{
		Order: entityToProto(entity),
	}, nil
}

func (s *GRPCServer) List(ctx context.Context, req *pb.ListOrderRequest) (*pb.ListOrderResponse, error) {
	entities, err := s.service.List(ctx)
	if err != nil {
		return nil, err
	}
	var items []*pb.Order
	for _, e := range entities {
		items = append(items, entityToProto(e))
	}
	return &pb.ListOrderResponse{
		Orders: items,
	}, nil
}

func (s *GRPCServer) Update(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}
	entity := &Order{
		ID:   id,
		Name: req.Name,
	}
	if err := s.service.Update(ctx, entity); err != nil {
		return nil, err
	}
	return &pb.UpdateOrderResponse{
		Order: entityToProto(entity),
	}, nil
}

func (s *GRPCServer) Delete(ctx context.Context, req *pb.DeleteOrderRequest) (*pb.DeleteOrderResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}
	if err := s.service.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &pb.DeleteOrderResponse{}, nil
}

func entityToProto(e *Order) *pb.Order {
	return &pb.Order{
		Id:   e.ID.String(),
		Name: e.Name,
	}
}
