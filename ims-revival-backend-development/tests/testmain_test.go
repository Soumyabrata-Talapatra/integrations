package tests

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"plirevival/bootstrap"
	"plirevival/routes"

	// bootstrapper "gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper"
	router "gitlab.cept.gov.in/it-2.0-common/n-api-server"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	db "gitlab.cept.gov.in/it-2.0-common/n-api-db"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

// Router holds the server router (for compatibility)
var Router *router.Router

// Engine is the gin.Engine used for HTTP test requests
var Engine *gin.Engine

var Fxconfig = fx.Module(
	"configmodule",
	fx.Provide(
		config.NewDefaultConfigFactory,
		newFxConfig,
	),
)

type FxConfigParam struct {
	fx.In
	Factory config.ConfigFactory
}

func newFxConfig(p FxConfigParam) (*config.Config, error) {
	return p.Factory.Create(
		config.WithFileName("config"),
		//config.WithAppEnv(os.Getenv("APP_ENV")),
		config.WithFilePaths(
			".",
			"../configs",
			//os.Getenv("APP_CONFIG_PATH"),
		),
	)
}

var FxDB = fx.Module(
	"DBModule",
	fx.Provide(
		SetUpDB,
	),
	// fx.Invoke(dblifecycle),
)

func SetUpDB(c *config.Config) (*db.DB, testcontainers.Container) {
	ctx := context.Background()
	var db1 *pgxpool.Pool
	var err error
	db1, Container, err = setupdockerdb(ctx, c)
	if err != nil {
		log.Fatal(ctx, "failed to setup db--->>> %s", err)
	}
	db := db.DB{Pool: db1}
	log.Info(ctx, "Successfully connected to the database %s", c.GetString("db.database"))
	return &db, Container
}

func setupdockerdb(ctx context.Context, c *config.Config) (*pgxpool.Pool, testcontainers.Container, error) {
	var env = map[string]string{
		"POSTGRES_PASSWORD": "password",
		"POSTGRES_USER":     "username",
		"POSTGRES_DB":       "database",
	}
	var port = "5432/tcp"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:14-alpine",
			ExposedPorts: []string{port},
			Env:          env,
			ShmSize:      150 * 1024 * 1024,
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, container, fmt.Errorf("failed to start container: %v", err)
	}

	p, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, container, fmt.Errorf("failed to get container external port: %v", err)
	}

	dbAddr := fmt.Sprintf("localhost:%s", p.Port())

	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s search_path=%s sslmode=disable",
		"username",
		"password",
		"localhost",
		p.Port(),
		"database",
		c.GetString("db.schema"))

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {

		return nil, container, err
	}
	config.MaxConns = int32(c.GetInt("db.maxconns"))
	config.MinConns = int32(c.GetInt("db.minconns"))
	config.MaxConnLifetime = time.Duration(c.GetInt("db.maxconnlifetime")) * time.Minute
	config.MaxConnIdleTime = time.Duration(c.GetInt("db.maxconnidletime")) * time.Minute
	retries := 0
	var db *pgxpool.Pool
	for retries < 10 {
		db, _ = pgxpool.New(ctx, config.ConnString())

		err := db.Ping(ctx)
		if err == nil {
			log.Info(ctx, "Ping Successful....")
			break
		} else {

			db.Close()
		}

		retries++
		log.Info(ctx, "Ping attempt failed. Retrying... (Attempt %d/%d)\n", retries, 10)
		time.Sleep(1 * time.Second)
	}

	err = migrateDb(dbAddr, c)
	if err != nil {
		log.Fatal(ctx, "failed to perform db migration--->>> %s", err)
	}

	return db, container, nil
}

func migrateDb(dbAddr string, c *config.Config) error {
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get path")
	}
	//pathToMigrationFiles := filepath.Dir(path) + "/migration"
	pathToMigrationFiles := filepath.Join(filepath.Dir(path), "migration")

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "username", "password", dbAddr, "database")

	m, err := migrate.New(fmt.Sprintf("file:%s", pathToMigrationFiles), databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func teardownTestData() {
	_ = Container.Terminate(context.Background())
	//	_ = MinioContainer.Terminate(context.Background())

	App.RequireStop()
}

var Container testcontainers.Container
var App *fxtest.App

// FxRouterParams defines the Fx dependencies for router creation
type FxRouterParams struct {
	fx.In
	Config   *config.Config
	Handlers []serverHandler.Handler `group:"servercontrollers"`
}

// NewTestRouter creates a gin.Engine and Router for testing
func NewTestRouter(p FxRouterParams) (*gin.Engine, *router.Router) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	// Setup the router framework
	router.Setup(engine)

	// Parse controllers and create registries
	registries := router.ParseControllers(p.Handlers...)

	// Create the router
	r := router.NewRouter(engine, p.Config, registries)

	// Register all routes
	r.RegisterRoutes()

	return engine, r
}

// FxRouter is the Fx module for router creation
var FxRouter = fx.Module(
	"RouterModule",
	fx.Provide(NewTestRouter),
)

// NewTestTemporalClient creates a mock Temporal client for testing
func NewTestTemporalClient(cfg *config.Config) (tclient.Client, error) {
	// For integration tests, connect to a real Temporal server
	// For unit tests, this can be mocked
	hostPort := cfg.GetString("temporal.host") + ":" + cfg.GetString("temporal.port")
	return tclient.Dial(tclient.Options{
		HostPort:  hostPort,
		Namespace: cfg.GetString("temporal.namespace"),
	})
}

// FxTemporal is the Fx module for Temporal client in tests
var FxTemporal = fx.Module(
	"TemporalTestModule",
	fx.Provide(NewTestTemporalClient),
)

func BootstrapTestApp(tb testing.TB, options ...fx.Option) *router.Router {
	tb.Helper()

	App = fxtest.New(
		tb,
		fx.Options(
			fx.Provide(func() context.Context {
				return context.Background()
			}),
		),
		Fxconfig,
		// bootstrapper.Fxlog,
		FxDB,
		// bootstrap.Fxvalidator, // Old validator - using govalid instead
		bootstrap.FxActivities,
		FxTemporal,
		bootstrap.FxHandler,
		bootstrap.FxRepo,
		//FxMinIO,
		FxRouter,
		fx.Populate(&Engine, &Router),
		fx.Invoke(routes.Routes),
	)
	App.RequireStart()
	return Router
}

func TestMain(m *testing.M) {
	t := &testing.T{}
	Router = BootstrapTestApp(t)
	m.Run()
	teardownTestData()
}

var MinioContainer *tcminio.MinioContainer

// Function to initialize MinIO
func newTestFxMinio(client *minio.Client, cfg *config.Config) error {
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.GetString("minio.bucketName"))
	if err != nil {
		log.Info(ctx, "Error checking if bucket exists")
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if exists {
		log.GetBaseLoggerInstance().ToZerolog().Debug().Msg("Bucket found")
	} else {
		err := client.MakeBucket(context.Background(), cfg.GetString("minio.bucketName"), minio.MakeBucketOptions{})
		if err != nil {
			log.Info(ctx, "Error creating bucket")
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Info(ctx, "Bucket created successfully")
	}
	return nil
}

// Function to set up MinIO test container
func SetUpMinio(ctx context.Context, cfg *config.Config) (*minio.Client, *tcminio.MinioContainer, error) {
	// Run the MinIO test container
	var err error
	MinioContainer, err = tcminio.Run(ctx,
		"minio/minio:RELEASE.2024-01-16T16-07-38Z")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start MinIO container: %w", err)
	}

	// Retrieve connection details
	url, err := MinioContainer.ConnectionString(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get MinIO connection string: %w", err)
	}

	// Create MinIO client
	minioClient, err := minio.New(url, &minio.Options{
		Creds:  credentials.NewStaticV4(MinioContainer.Username, MinioContainer.Password, ""),
		Secure: false, // Use false for testing; set true in production with HTTPS
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}
	bucketName := cfg.GetString("minio.bucketName")

	if err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
		log.Info(ctx, "Failed to create bucket: %v", err)
	}

	return minioClient, MinioContainer, nil
}

// Fx Module for MinIO
var FxMinIO = fx.Module(
	"MinIOModule",
	fx.Provide(func(ctx context.Context, cfg *config.Config) (*minio.Client, *tcminio.MinioContainer, error) {
		return SetUpMinio(ctx, cfg)
	}),
	fx.Invoke(newTestFxMinio),
)
