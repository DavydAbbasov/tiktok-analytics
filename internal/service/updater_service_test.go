package service

import (
	"context"
	"fmt"
	"testing"
	"time"
	"ttanalytic/internal/mocks"
	"ttanalytic/internal/models"

	"github.com/golang/mock/gomock"
)

func TestUpdaterService_processBatch_NoVideos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUpdaterRepository(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	logger.EXPECT().
		Info("updater: no videos to update").
		Times(1)

	cfg := UpdaterConfig{
		Interval:       time.Second,
		BatchSize:      10,
		MinUpdateAge:   time.Second,
		MaxConcurrency: 1,
	}
	earningsCfg := EarningsConfig{Per: 1000, Rate: 0.10}

	u := NewUpdaterService(
		repo,
		nil,
		logger,
		cfg,
		earningsCfg,
		nil,
	)

	ctx := context.Background()

	repo.EXPECT().
		ListVideosForUpdate(gomock.Any(), cfg.MinUpdateAge, cfg.BatchSize).
		Return([]models.Video{}, nil)

	if err := u.processBatch(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdaterService_processBatch_SingleBatchTwoVideos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUpdaterRepository(ctrl)
	provider := mocks.NewMockTikTokProvider(ctrl)
	logger := mocks.NewMockLogger(ctrl)
	transactor := mocks.NewMockTransactor(ctrl)

	logger.EXPECT().Info(gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	cfg := UpdaterConfig{
		Interval:       time.Second,
		BatchSize:      10,
		MinUpdateAge:   0,
		MaxConcurrency: 1,
	}
	earningsCfg := EarningsConfig{Per: 1000, Rate: 0.10}

	u := NewUpdaterService(repo, provider, logger, cfg, earningsCfg, transactor)

	ctx := context.Background()

	videos := []models.Video{
		{ID: 1, URL: "url1", TikTokID: "t1", CurrentViews: 0, CurrentEarnings: 0},
		{ID: 2, URL: "url2", TikTokID: "t2", CurrentViews: 0, CurrentEarnings: 0},
	}

	stats1 := &models.VideoStats{Views: 100}
	stats2 := &models.VideoStats{Views: 200}

	gomock.InOrder(
		repo.EXPECT().
			ListVideosForUpdate(gomock.Any(), cfg.MinUpdateAge, cfg.BatchSize).
			Return(videos, nil),
		repo.EXPECT().
			ListVideosForUpdate(gomock.Any(), cfg.MinUpdateAge, cfg.BatchSize).
			Return([]models.Video{}, nil),
	)

	provider.EXPECT().
		GetVideoStats(gomock.Any(), "url1").
		Return(stats1, nil)
	provider.EXPECT().
		GetVideoStats(gomock.Any(), "url2").
		Return(stats2, nil)

	repo.EXPECT().
		AppendVideoStats(gomock.Any(), gomock.Any()).
		Times(2).
		Return(nil)
	repo.EXPECT().
		UpdateVideoAggregates(gomock.Any(), gomock.Any()).
		Times(2).
		Return(nil)

	transactor.EXPECT().
		WithinTransaction(gomock.Any(), gomock.Any()).
		Times(2).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})

	if err := u.processBatch(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
func TestUpdaterService_processBatch_MultipleBatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//mock environment
	repo := mocks.NewMockUpdaterRepository(ctrl)
	provider := mocks.NewMockTikTokProvider(ctrl)
	logger := mocks.NewMockLogger(ctrl)
	transactor := mocks.NewMockTransactor(ctrl)

	//mock logger
	logger.EXPECT().Info(gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	cfg := UpdaterConfig{
		Interval:       time.Second,
		BatchSize:      10, //limit
		MinUpdateAge:   0,
		MaxConcurrency: 1,
	}
	//formula
	earningsCfg := EarningsConfig{
		Per:  1000,
		Rate: 0.10,
	}

	u := NewUpdaterService(repo, provider, logger, cfg, earningsCfg, transactor)
	ctx := context.Background()

	//create 20 video in db
	var allVideos []models.Video
	for i := 1; i <= 21; i++ {
		allVideos = append(allVideos, models.Video{
			ID:              int64(i),
			URL:             fmt.Sprintf("url-%d", i),
			TikTokID:        fmt.Sprintf("id tt-%d", i),
			CurrentViews:    0,
			CurrentEarnings: 0,
		})
	}

	firstBatch := allVideos[:10]    // 0..9  -> 1..10
	secondBatch := allVideos[10:20] // 10..19 -> 11..20
	lastBatch := allVideos[20:]

	//Expectations
	gomock.InOrder(
		repo.EXPECT().
			ListVideosForUpdate(gomock.Any(), cfg.MinUpdateAge, cfg.BatchSize).
			Return(firstBatch, nil),
		repo.EXPECT().
			ListVideosForUpdate(gomock.Any(), cfg.MinUpdateAge, cfg.BatchSize).
			Return(secondBatch, nil),
		repo.EXPECT().
			ListVideosForUpdate(gomock.Any(), cfg.MinUpdateAge, cfg.BatchSize).
			Return(lastBatch, nil),
		repo.EXPECT().
			ListVideosForUpdate(gomock.Any(), cfg.MinUpdateAge, cfg.BatchSize).
			Return([]models.Video{}, nil),
	)

	for _, v := range allVideos {
		stats := &models.VideoStats{
			Views: v.ID * 10,
		}
		provider.EXPECT().
			GetVideoStats(gomock.Any(), v.URL).
			Return(stats, nil)
		// AnyTimes() //bloc concurency
	}

	//work with relationships
	repo.EXPECT().
		//append new video stats
		AppendVideoStats(gomock.Any(), gomock.Any()).
		Times(21). //>20
		Return(nil)

	repo.EXPECT().
		//update arrguments
		UpdateVideoAggregates(gomock.Any(), gomock.Any()).
		Times(21). //>20
		Return(nil)

	//corect work with transaction
	transactor.EXPECT().
		//call transaction every time when we call repo
		WithinTransaction(gomock.Any(), gomock.Any()).
		Times(21). //>20
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})

	err := u.processBatch(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
