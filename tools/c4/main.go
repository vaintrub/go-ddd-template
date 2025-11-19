package main

import (
	"context"
	"fmt"
	"os"

	"github.com/krzysztofreczek/go-structurizr/pkg/scraper"
	"github.com/krzysztofreczek/go-structurizr/pkg/view"

	"github.com/vaintrub/go-ddd-template/internal/common/config"
	trainerService "github.com/vaintrub/go-ddd-template/internal/trainer/service"
	trainingsService "github.com/vaintrub/go-ddd-template/internal/trainings/service"
)

const (
	scraperConfig = "scraper.yml"
	viewConfig    = "view.yml"
	outputFile    = "out/view-%s.plantuml"
)

func main() {
	ctx := context.Background()

	cfg := config.MustLoad(ctx)

	trainerApp := trainerService.NewApplication(ctx, cfg)
	scrape(trainerApp, "trainer")

	trainingsApp, _ := trainingsService.NewApplication(ctx, cfg)
	scrape(trainingsApp, "trainings")
}

func scrape(app interface{}, name string) {
	s, err := scraper.NewScraperFromConfigFile(scraperConfig)
	if err != nil {
		panic(err)
	}

	structure := s.Scrape(app)

	outFileName := fmt.Sprintf(outputFile, name)
	outFile, err := os.Create(outFileName)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = outFile.Close()
	}()

	v, err := view.NewViewFromConfigFile(viewConfig)
	if err != nil {
		panic(err)
	}

	err = v.RenderStructureTo(structure, outFile)
	if err != nil {
		panic(err)
	}
}
