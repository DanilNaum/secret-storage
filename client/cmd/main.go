package main

import (
	"log"

	"github.com/DanilNaum/secret-storage-client/internal/constants"
	"github.com/DanilNaum/secret-storage-client/internal/localhandlers"
	"github.com/DanilNaum/secret-storage-client/internal/models"
	"github.com/DanilNaum/secret-storage-client/internal/serverhandlers"
	"github.com/DanilNaum/secret-storage-client/internal/storage"
	"github.com/DanilNaum/secret-storage-client/internal/ui"
	"github.com/DanilNaum/secret-storage-client/pkg/cache"
	"github.com/rivo/tview"
)

func main() {
	stor := storage.NewStorage()

	serverCache := cache.NewCache[models.Record](constants.DefaultCacheTTL)
	
	localHandlers := localhandlers.NewLocalHandlers(stor)
	
	serverURL := "localhost:9090"
	
	serverHandlers, err := serverhandlers.NewServerHandlers(stor, serverCache, serverURL)
	if err != nil {
		log.Fatal(err)
	}
	
	app := ui.NewApp(stor, localHandlers, serverHandlers, tview.NewPages())
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}