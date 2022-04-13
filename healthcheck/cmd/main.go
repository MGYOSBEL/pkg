package main

import (
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/MGYOSBEL/pkg/healthcheck"
)

type DbClient struct {
	server string
}

func (db DbClient) Check() (bool, error) {
	status := randomStatus()
	return status, nil
}

type MqttClient struct {
	server string
}

func (mqtt MqttClient) Check() (bool, error) {
	status := randomStatus()
	return status, nil
}

func main() {
	// sugar zap logger
	config := zap.NewProductionEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.TimeKey = "time"
	config.EncodeCaller = zapcore.ShortCallerEncoder
	config.EncodeTime = zapcore.TimeEncoderOfLayout("02/01/2006 15:04:05")
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(config),
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.DebugLevel,
	))
	defer logger.Sync()
	sugar := logger.Sugar()

	checker := healthcheck.New("health", logger)

	mqtt := MqttClient{
		server: "localhost",
	}
	db := DbClient{
		server: "localhost",
	}

	sugar.Infof("Registering checkers")
	go func() {
		checker.Register("database", db)
		checker.Register("mqtt", mqtt)
		_ = http.ListenAndServe(":8080", nil)
	}()

	Run(logger)
}

func randomStatus() bool {
	return rand.Int()%2 == 0
}

var (
	sugar *zap.SugaredLogger
	// cfg   *config.Config
)

func Run(logger *zap.Logger) {
	// cfg = config.LoadConfig()
	sugar = logger.Sugar()
	go Execute()

	WaitSignal()

}

func Execute() {
	for {
		sugar.Infof("🔥🔥🔥 Startign test 🔥🔥🔥")
		time.Sleep(1 * time.Minute)
		sugar.Infof("✅ Test succesfully ended")
	}
}

func WaitSignal() os.Signal {
	ch := make(chan os.Signal, 2)
	signal.Notify(
		ch,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
			return sig
		}
	}
}
