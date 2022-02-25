package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	dev = kingpin.Flag("dev", "Режим development.").Short('d').Bool()
	env = kingpin.Arg("env", "Конфигурация приложения.").Default(".env").String()

	db *leveldb.DB
	// Состояние сервиса
	serviceStatus *serviceStatusInfo
)

func main() {
	var (
		err error
	)

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	if err = godotenv.Load(*env); err != nil {
		fmt.Println("Error loading env file")
		os.Exit(1)
	}

	processName := filepath.Base(os.Args[0])

	// Создадим лог-файл
	logFile := &lumberjack.Logger{
		Filename:   "logs/" + processName + ".log",
		MaxSize:    maxLogfile,
		MaxBackups: MaxBackupsLog,
		MaxAge:     maxAgeLogfile,
		Compress:   true,
	}
	defer func() {
		if derr := logFile.Close(); derr != nil {
			log.Errorf("Ошибка закрытия лог-файла %s %s", logFile.Filename, derr)
		}
	}()

	if *dev {
		log.SetOutput(os.Stdout)

	} else {
		log.SetLevel(log.WarnLevel)
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(logFile)
	}

	db, err = leveldb.OpenFile(os.Getenv("DB_PATH"), nil)
	defer func(db *leveldb.DB) {
		if derr := db.Close(); derr != nil {
			log.Println(derr)
		}
	}(db)

	serviceStatus = newServiceStatusInfo()
	signal := make(chan string, 1)

	m := NewMiddleware(NewRouting(db), signal)
	serv := &http.Server{
		Addr:        os.Getenv("ADDR"),
		Handler:     m,
		ReadTimeout: 5 * time.Second,
		//ReadHeaderTimeout: 0,
		WriteTimeout: 10 * time.Second,
	}

	// Запускаем http сервер
	log.Printf("Serving http on -addr=%q", os.Getenv("ADDR"))
	go func() {
		if err = serv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error in ListenAndServe: %s", err)
		}
	}()

loopSignal:
	for s := range signal {
		switch s {
		case "stop":
			break loopSignal
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = serv.Shutdown(ctx); err != nil {
		fmt.Printf("error in ListenAndServe: %s", err)
		os.Exit(1)
	}
}
