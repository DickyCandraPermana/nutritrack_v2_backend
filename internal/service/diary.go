package service

import (
	"context"
	"errors"
	"time"

	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/mapper"
	"github.com/MyFirstGo/internal/store"
	"github.com/go-playground/validator/v10"
	"golang.org/x/sync/errgroup"
)

type DiaryService struct {
	store     store.Storage
	validator validator.Validate
}

func (s *DiaryService) GetSummaryByUserId(ctx context.Context, userID int64, date time.Time) (*domain.DailySummary, error) {
	var summary *domain.DailySummary
	var entries []*domain.FoodDiary

	// 1. Inisialisasi errgroup dengan context
	g, ctx := errgroup.WithContext(ctx)

	// 2. Jalankan fungsi GetSummary secara konkuren
	g.Go(func() error {
		var err error
		summary, err = s.store.Diary.GetSummary(ctx, userID, date)
		return err // Error ini akan ditangkap oleh g.Wait()
	})

	// 3. Jalankan fungsi GetEntries secara konkuren
	g.Go(func() error {
		var err error
		entries, err = s.store.Diary.GetEntries(ctx, userID, date)
		return err
	})

	// 4. Tunggu semua goroutine selesai dan cek apakah ada yang error
	if err := g.Wait(); err != nil {
		return nil, err // Jika salah satu error, kita kembalikan error tersebut
	}

	// 5. Gabungkan data setelah keduanya sukses
	if summary != nil && len(entries) > 0 {
		// Pastikan slice diinisialisasi
		if summary.Entries == nil {
			summary.Entries = make([]domain.FoodDiary, 0, len(entries))
		}

		for _, entry := range entries {
			summary.Entries = append(summary.Entries, *entry)
		}
	}

	return summary, nil
}

func (s *DiaryService) GetDiaryWithUserId(ctx context.Context, userID, diaryID int64) (*domain.FoodDiary, error) {
	diary, err := s.store.Diary.GetUserEntry(ctx, userID, diaryID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return diary, nil
}

func (s *DiaryService) GetDiaryByDiaryId(ctx context.Context, diaryID int64) (*domain.FoodDiary, error) {

	diary, err := s.store.Diary.GetEntry(ctx, diaryID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return diary, nil
}

func (s *DiaryService) Create(ctx context.Context, input *domain.DiaryCreateInput) (*domain.FoodDiary, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	if input.ConsumedAt.IsZero() {
		input.ConsumedAt = time.Now()
	}

	diary := mapper.CreateDiaryInputToFoodDiary(input)

	err := s.store.Diary.Create(ctx, diary)
	if err != nil {
		return nil, err
	}

	return diary, nil
}

func (s *DiaryService) Update(ctx context.Context, userID int64, input *domain.DiaryUpdateInput) (*domain.FoodDiary, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	diary, err := s.GetDiaryWithUserId(ctx, userID, input.ID)
	if err != nil {
		return nil, err
	}

	if input.AmountConsumed != nil {
		diary.AmountConsumed = *input.AmountConsumed
	}

	if input.ConsumedAt != nil {
		diary.ConsumedAt = *input.ConsumedAt
	}

	if input.MealType != nil {
		diary.MealType = *input.MealType
	}

	err = s.store.Diary.Update(ctx, diary)
	if err != nil {
		return nil, err
	}

	return diary, nil
}

func (s *DiaryService) Delete(ctx context.Context, userID, diaryID int64) error {
	_, err := s.GetDiaryWithUserId(ctx, userID, diaryID)
	if err != nil {
		return err
	}

	if err = s.store.Diary.Delete(ctx, diaryID); err != nil {
		return err
	}

	return nil
}
