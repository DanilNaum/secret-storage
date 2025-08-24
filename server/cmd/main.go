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

	// Выносим инициализацию зависимостей в отдельные функции
	grpcServer, cleanup, err := initDependencies(ctx, conf)
	if err != nil {
		return fmt.Errorf("failed to initialize dependencies: %w", err)
	}
	defer cleanup()

	grpcServer.Run(ctx)

	err = <-grpcServer.Notify()
	if err != nil {
		return fmt.Errorf("grpc server failed with error %w", err)
	}
	return nil
}

// initDependencies инициализирует все зависимости приложения
func initDependencies(ctx context.Context, conf *config.Config) (*grpcserver.GrpcServer, func(), error) {
	// Инициализация миграций
	if err := runMigrations(conf); err != nil {
		return nil, nil, err
	}

	// Инициализация хранилища
	storage, cleanupStorage, err := initStorage(ctx, conf)
	if err != nil {
		return nil, nil, err
	}

	// Инициализация утилит
	jwtManager, hasher := initUtilities(conf)

	// Инициализация auth слоя
	authHandler := initAuthLayer(storage, jwtManager, hasher)

	// Инициализация record слоя
	recordHandler := initRecordLayer(storage, jwtManager)

	// Инициализация GRPC сервера
	grpcServer := grpcserver.NewGrpcServer(conf.GetGRPCPort(), authHandler, recordHandler)

	// Функция для очистки ресурсов
	cleanup := func() {
		cleanupStorage()
	}

	return grpcServer, cleanup, nil
}

// runMigrations выполняет миграции базы данных
func runMigrations(conf *config.Config) error {
	migrator := migration.NewMigrator(conf.GetDSN(), migration.WithRelativePath("migrations"))
	if err := migrator.Migrate(); err != nil {
		return fmt.Errorf("migration failed with error %w", err)
	}
	return nil
}

// initStorage инициализирует подключение к базе данных и хранилище
func initStorage(ctx context.Context, conf *config.Config) (*storage.Storage, func(), error) {
	pgConf := pg.NewConnConfigFromDsnString(conf.GetDSN())
	pgConn, err := pg.NewConnection(ctx, pgConf)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	storage := storage.NewStorage(pgConn)

	cleanup := func() {
		if pgConn != nil {
			pgConn.Close()
		}
	}

	return storage, cleanup, nil
}

// initUtilities инициализирует вспомогательные утилиты
func initUtilities(conf *config.Config) (*jwt.JWTManager[int], *hasher.Argon2Hasher) {
	jwtManager := jwt.NewJWTManager[int](conf.GetJWTSecret(), conf.GetJWTDuration())
	hasher := hasher.NewArgon2Hasher(nil)
	return jwtManager, hasher
}

// initAuthLayer инициализирует слой аутентификации
func initAuthLayer(
	storage *storage.Storage,
	jwtManager *jwt.JWTManager[int],
	hasher *hasher.Argon2Hasher,
) *authhandler.GrpcAuthHandler {
	authUsecase := auth.NewAuthUsecase(storage, jwtManager, hasher)
	return authhandler.NewGrpcAuthHandler(authUsecase)
}

// initRecordLayer инициализирует слой работы с записями
func initRecordLayer(
	storage *storage.Storage,
	jwtManager *jwt.JWTManager[int],
) *recordhandler.GrpcRecordHandler {
	recordUsecase := record.NewRecordUsecase(storage, storage, storage)
	fileRepo := filerepository.NewFileRepository("./files")
	return recordhandler.NewGrpcRecordHandler(jwtManager, recordUsecase, fileRepo)
}
