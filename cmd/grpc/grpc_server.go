package main

import (
	"avito-backend-trainee-assignment-spring-2025/cmd/grpc/pvz_v1"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/interfaces"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/postgres"
	"avito-backend-trainee-assignment-spring-2025/pkg/config"
	"avito-backend-trainee-assignment-spring-2025/pkg/logger"
	"context"
	"fmt"
	"google.golang.org/grpc/reflection"
	"net"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PVZGrpcServer struct {
	pvz_v1.UnimplementedPVZServiceServer
	pvzRepo interfaces.TxPVZRepository
}

func (s *PVZGrpcServer) GetPVZList(ctx context.Context, req *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {
	log.Info().Msg("GRPC request: GetPVZList")

	filter := models.PVZFilter{
		Page:  1,
		Limit: 1000000000,
	}

	pvzList, _, err := s.pvzRepo.GetAll(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get PVZ list in GRPC handler")
		return nil, err
	}

	var protoPVZs []*pvz_v1.PVZ
	for _, pvz := range pvzList {
		protoPVZ := &pvz_v1.PVZ{
			Id:               pvz.ID.String(),
			RegistrationDate: timestamppb.New(pvz.RegistrationDate),
			City:             pvz.City,
		}
		protoPVZs = append(protoPVZs, protoPVZ)
	}

	return &pvz_v1.GetPVZListResponse{
		Pvzs: protoPVZs,
	}, nil
}

func StartGRPCServer(cfg *config.Config, pvzRepo interfaces.TxPVZRepository) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pvzService := &PVZGrpcServer{
		pvzRepo: pvzRepo,
	}

	reflection.Register(grpcServer)

	pvz_v1.RegisterPVZServiceServer(grpcServer, pvzService)

	log.Info().Str("port", cfg.GRPC.Port).Msg("Starting gRPC server")
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err.Error())
		return
	}

	logger.Setup(cfg.Logger)

	db, err := postgres.New(&cfg.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	pvzRepo := postgres.NewPVZRepository(db)

	if err := StartGRPCServer(cfg, pvzRepo); err != nil {
		log.Fatal().Err(err).Msg("Failed to start gRPC server")
	}
}
