package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/syndtr/goleveldb/leveldb"
	"tfacoinlist/route"
)

const (
	homePage = "/"
)

func NewRouting(db *leveldb.DB) *httprouter.Router {
	router := httprouter.New()
	// Страница для ручного добавления аккаунта
	router.GET("/manual-registration/", route.ManualRegistrationGET())
	// Ручная регистрация
	router.POST("/manual-registration/", route.ManualRegistration(db))

	// Статус сервера
	router.GET(homePage, route.Homepage())
	// Регистрация
	router.POST("/registration/", route.Registration(db))

	authRoutes(router, db)
	return router
}

// authRoutes маршруты для работы с кодами авторизации
func authRoutes(router *httprouter.Router, db *leveldb.DB) {
	// Получение кода
	router.GET("/auth/totp/:email/:key/", route.AuthCode(db))
}
