package bootstrap

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	"gitlab.cept.gov.in/it-2.0-common/cache"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"

	handler "plirevival/handler"
	repo "plirevival/repo/postgres"
	workflow "plirevival/workflow"

	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	"go.uber.org/fx"
)

// NewValidatorService add it as part of fx invoke
var Fxvalidator = fx.Module(
	"validator",
	fx.Invoke(handler.NewValidatorService),
)

var FxRepo = fx.Module(
	"Repomodule",
	fx.Provide(
		// repo.NewAwardsRepository,

		// repo.NewCommunicationRepository,

		// repo.NewDocRepository,
		// repo.NewUserRepository,

		repo.NewRevivalRepository,
		repo.NewPolicyRepository,
		repo.NewPaymentRepository,
	),
)

var FxActivities = fx.Module(
	"ActivitiesModule",
	fx.Provide(
		workflow.NewActivities,
	),
)

var FxHandler = fx.Module(
	"Handlermodule",
	fx.Provide(

		// fx.Annotate(
		// 	handler.NewDocHandler,
		// 	fx.As(new(serverHandler.Handler)),
		// 	fx.ResultTags(serverHandler.ServerControllersGroupTag),
		// ),
		// fx.Annotate(
		// 	handler.NewUserHandler,
		// 	fx.As(new(serverHandler.Handler)),
		// 	fx.ResultTags(serverHandler.ServerControllersGroupTag),
		// ),
		fx.Annotate(
			handler.NewRevivalHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
	),
)

var FxCache = fx.Module(
	"Cachemodule",
	fx.Provide(CacheInit),
)

func CacheInit(c *config.Config) cache.CacheInterface {
	redisServer := c.GetString("cache.redisserver")
	redisPassword := c.GetString("cache.redispassword")
	redisDBIndex := c.GetInt("cache.redisdbindex")
	QueryTimeoutMed := c.GetDuration("db.QueryTimeoutMed")
	redisTTL := c.GetDuration("cache.redisexpirationtime")
	lcCapacity := c.GetInt("cache.lccapacity")
	lcNumShards := c.GetInt("cache.lcnumshards")
	lcBatchSize := c.GetInt("cache.lcbatchsize")
	lcBatchBufferTimeout := c.GetDuration("cache.lcbatchbuffertimeout")
	lcEvictionPercentage := c.GetInt("cache.lcevictionpercentage")
	lcMaxRefreshDelay := c.GetDuration("cache.lcmaxrefreshdelay")
	lcMinRefreshDelay := c.GetDuration("cache.lcminrefreshdelay")
	lcRetryBaseDelay := c.GetDuration("cache.lcretrybasedelay")
	lcTTL := c.GetDuration("cache.lcttl")
	isRedisCacheEnabled := c.GetBool("cache.isredisenabled")
	isLocalCacheEnabled := c.GetBool("cache.islocalcacheenabled")
	cachePointer, err := cache.NewCache(
		redisServer, redisPassword, redisDBIndex, QueryTimeoutMed, redisTTL, lcCapacity,
		lcNumShards, lcBatchSize, lcBatchBufferTimeout,
		lcEvictionPercentage, lcMaxRefreshDelay, lcMinRefreshDelay,
		lcRetryBaseDelay, lcTTL, isRedisCacheEnabled, isLocalCacheEnabled,
	)

	if err != nil {
		log.Warn(nil, "Error: %v", err)
	} else {
		log.Info(nil, "Info: Local Cache and Redis Cache are Initialized")
	}
	if !isRedisCacheEnabled {
		log.Info(nil, "Info: Redis Cache is disabled")
	}
	if !isLocalCacheEnabled {
		log.Info(nil, "Info: Local Cache is disabled")
	}
	return cachePointer
}

func InitMinio(c *config.Config) *minio.Client {
	var err error
	MinioClient, err := minio.New(c.GetString("minio.url"), &minio.Options{
		Creds:  credentials.NewStaticV4(c.GetString("minio.accessKey"), c.GetString("minio.secretKey"), ""),
		Secure: true,
	})
	if err != nil {
		log.Fatal(nil, err)
	}

	exists, errBucketExists := MinioClient.BucketExists(context.Background(), c.GetString("minio.bucketName"))

	if errBucketExists != nil {
		log.Fatal(nil, "Error checking if bucket exists: %v", errBucketExists)
	}

	if exists {
		fmt.Println("Bucket found")
	} else {
		log.Fatal(nil, "Bucket %s does not exist", c.GetString("minio.bucketName"))
	}
	return MinioClient
}
