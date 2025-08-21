package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DanilNaum/secret-storage-server/internal/config"
	"github.com/DanilNaum/secret-storage-server/internal/filerepository"
	"github.com/DanilNaum/secret-storage-server/internal/storage"
	authhandler "github.com/DanilNaum/secret-storage-server/internal/transport/grpc/auth/v1"
	recordhandler "github.com/DanilNaum/secret-storage-server/internal/transport/grpc/record/v1"
	"github.com/DanilNaum/secret-storage-server/internal/usecase/auth"
	"github.com/DanilNaum/secret-storage-server/internal/usecase/record"
	"github.com/DanilNaum/secret-storage-server/pkg/grpcserver"
	"github.com/DanilNaum/secret-storage-server/pkg/hasher"
	"github.com/DanilNaum/secret-storage-server/pkg/jwt"
	"github.com/DanilNaum/secret-storage-server/pkg/migration"
	"github.com/DanilNaum/secret-storage-server/pkg/pg"
)

func main() {
	log := log.New(os.Stdout, "secret-storage-server: ", log.LstdFlags)
	err := run(log)
	if err != nil {
		log.Fatal(err)
	}

}

func run(log *log.Logger) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	conf, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("getting config failed with err: %w", err)
	}
	if conf.GetDSN() == "" {
		return errors.New("dsn is empty")
	}

	migrator := migration.NewMigrator(conf.GetDSN(), migration.WithRelativePath("migrations"))
	err = migrator.Migrate()
	if err != nil {
		return fmt.Errorf("migration failed with error %w", err)
	}

	pgConf := pg.NewConnConfigFromDsnString(conf.GetDSN())
	pgConn, err := pg.NewConnection(ctx, pgConf)
	storage := storage.NewStorage(pgConn)

	jwtManager := jwt.NewJWTManager[int](conf.GetJWTSecret(), conf.GetJWTDuration())
	hasher := hasher.NewArgon2Hasher(nil)

	authUsecase := auth.NewAuthUsecase(storage, jwtManager, hasher)
	authHandler := authhandler.NewGrpcAuthHandler(authUsecase)

	recordUsecase := record.NewRecordUsecase(storage, storage, storage)
fileRepo := filerepository.NewFileRepository("./files")

	recordHandler := recordhandler.NewGrpcRecordHandler(jwtManager, recordUsecase, fileRepo)
	grpcserver := grpcserver.NewGrpcServer(conf.GetGRPCPort(), authHandler, recordHandler)
	grpcserver.Run(ctx)

	err = <-grpcserver.Notify()
	if err != nil {
		return fmt.Errorf("grpc server failed with error %w", err)
	}
	return nil
}
