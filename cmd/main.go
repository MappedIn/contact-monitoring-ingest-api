package main

import (
	"contact-monitoring-ingest-api/internal/auth"
	"contact-monitoring-ingest-api/internal/device"
	"contact-monitoring-ingest-api/internal/health"
	"contact-monitoring-ingest-api/internal/invitecode"
	"contact-monitoring-ingest-api/internal/positionevent"
	"context"
	"hash/maphash"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload" // load environment vars from .env
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var port = os.Getenv("PORT")
var mongoURL = os.Getenv("MONGO_URL")
var mongoDBName = os.Getenv("MONGO_DB_NAME")
var deviceTokenSecret = os.Getenv("DEVICE_TOKEN_SECRET")
var inviteCodeUser = os.Getenv("INVITE_CODE_USER")
var inviteCodePass = os.Getenv("INVITE_CODE_PASS")
var maxDistanceBetweenDevices = os.Getenv("MAXIMUM_DISTANCE_BETWEEN_DEVICES")
var accThreshold = os.Getenv("ACCURACY_THRESHOLD")

func main() {
	log.Printf("PID: %d GOMAXPROCS is %d\n", os.Getpid(), runtime.GOMAXPROCS(0))

	var mongoDBMaxPoolSize uint64 = 100
	totalWorkers := 100 // total workers should be close to mongoDBMaxPoolSize

	if port == "" {
		log.Fatal("You must provide a PORT env")
	}

	if mongoURL == "" {
		log.Fatal("You must provide a MONGO_URL env")
	}

	if mongoDBName == "" {
		log.Fatal("You must provide a MONGO_DB_NAME env")
	}

	if deviceTokenSecret == "" {
		log.Fatal("You must provide a DEVICE_TOKEN_SECRET env")
	}

	if inviteCodeUser == "" {
		log.Fatal("You must provide a INVITE_CODE_USER env")
	}

	if inviteCodePass == "" {
		log.Fatal("You must provide a INVITE_CODE_PASS env")
	}

	var maximumDistanceBetweenDevices float64 = 0
	if maxDistanceBetweenDevices != "" {
		var err error
		maximumDistanceBetweenDevices, err = strconv.ParseFloat(maxDistanceBetweenDevices, 64)
		if err != nil {
			log.Fatal(err)
		}
	}

	var accuracyThreshold float64 = 0
	if accThreshold != "" {
		var err error
		accuracyThreshold, err = strconv.ParseFloat(accThreshold, 64)
		if err != nil {
			log.Fatal(err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// MONGO
	// we want the write concern to be "write majority" so that after
	// inserting a new position event we can be confident that
	// processing done afterwards will include it in subsequent queries.
	// If we don't have this write concern we can have cases where
	// concurrent inserts for positions near each other don't generate
	// a minute aggregate because both are missed in their nearby query
	log.Println("Connecting to mongo")
	dbOptions := options.Client().
		ApplyURI(mongoURL).
		SetMaxPoolSize(mongoDBMaxPoolSize).
		SetWriteConcern(writeconcern.New(writeconcern.WMajority()))

	dbClient, err := mongo.Connect(ctx, dbOptions)
	if err != nil {
		log.Fatal("Cannot connect to mongodb", err)
	}
	defer dbClient.Disconnect(ctx)

	// Call Ping to verify that the deployment is up and the Client was configured successfully.
	// As mentioned in the Ping documentation, this reduces application resiliency as the server may be
	// temporarily unavailable when Ping is called.
	go func() {
		if err = dbClient.Ping(ctx, readpref.Primary()); err != nil {
			log.Fatal(err)
		} else {
			log.Println("Connected to mongo successfully")
		}
	}()

	db := dbClient.Database(mongoDBName)
	eventCollection := db.Collection("position-event")

	eventChan := make(chan positionevent.PositionEvent, totalWorkers)
	minAggregateChan := make(chan positionevent.MinuteAggregate, totalWorkers)

	gin.DefaultWriter = os.Stdout

	codeRepo := invitecode.NewRepo(db.Collection("invite-code"))
	deviceRepo := device.NewRepo(db.Collection("device"))

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.GET("/health", health.GetHandler())

	router.POST(
		"/positions",
		auth.DeviceTokenMiddleware(deviceTokenSecret),
		positionevent.PostHandler(eventCollection, eventChan, accuracyThreshold),
	)

	inviteCodeRoutes := router.Group(
		"/invite-code",
		gin.BasicAuth(gin.Accounts{
			inviteCodeUser: inviteCodePass,
		}),
	)
	{
		inviteCodeRoutes.POST(":venue", invitecode.CreateHandler(codeRepo))
	}

	deviceRoutes := router.Group("/device")
	{
		deviceByIDRoutes := deviceRoutes.Group(":id")
		{
			deviceByIDRoutes.POST("activate", device.ActivateHandler(codeRepo, deviceRepo))
			deviceByIDRoutes.GET("token", auth.GetDeviceTokenHandler(deviceTokenSecret, deviceRepo))
		}
	}

	var wg sync.WaitGroup
	for i := 1; i <= totalWorkers; i++ {
		wg.Add(1)
		go positionevent.EventWorker(positionevent.EventWorkerConfig{
			DB:                            db,
			EventChan:                     eventChan,
			MinAggregateChan:              minAggregateChan,
			WG:                            &wg,
			MaximumDistanceBetweenDevices: maximumDistanceBetweenDevices,
			AccuracyThreshold:             accuracyThreshold,
			Name:                          i,
		})
	}

	partitions := make(map[int](chan positionevent.MinuteAggregate))
	for i := 0; i < totalWorkers; i++ {
		partitions[i] = make(chan positionevent.MinuteAggregate, 10)
		// along with the one goroutine that triages each event, each worker goroutine
		// represents a consumer on the minAggregateChan
		wg.Add(1)
		go positionevent.AggregateWorker(db, partitions[i], &wg, i)
	}

	go func() {
		var h maphash.Hash

		for minAggregate := range minAggregateChan {
			// using maphash to imitate partitioning of data on the person id pairs
			// this partitioning is required so we don't run into concurrency issues
			// (eg. concurrent upserts to the same records results in current inserts;
			// concurrent contact event writes for the same person id pair happen sequentially)
			h.Reset()
			h.WriteString(minAggregate.Events[0].DeviceID)
			h.WriteString(minAggregate.Events[1].DeviceID)
			partition := h.Sum64() % uint64(totalWorkers)
			// one problem is this will be blocking if the partition fills up
			partitions[int(partition)] <- minAggregate
		}
	}()

	// declare server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// start server in goroutine so we can handle graceful shutdown
	go func() {
		log.Println("Starting server")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("listen: %s\n", err)
		}
	}()

	// listen for signal from OS to kill this process
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

	// block until signal is received
	sig := <-sigChan
	log.Println("received kill signal:", sig)

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// gracefully attempt to shutdown server
	err = server.Shutdown(ctx)
	if err != nil {
		log.Fatalf("Server shutdown error %s:", err)
		os.Exit(1)
	}

	close(eventChan)
	close(minAggregateChan)
	for _, partitionChannel := range partitions {
		close(partitionChannel)
	}

	wg.Wait()
	log.Println("Server shutdown finished")
	os.Exit(0)
}
