package usecase

import (
	"context"

	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/leaderboard/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/leaderboard/repository"
)

type LeaderboardUseCase interface {
	AddPoints(ctx context.Context, req dto.PointRequest) error
	GetLeaderboard(ctx context.Context, serviceID string, limit int64) (dto.LeaderboardResponse, error)
	GetPartnerRank(ctx context.Context, userID, serviceID string) (dto.LeaderboardEntry, error)
}

type leaderboardUseCase struct {
	repo repository.LeaderboardRepository
}

func NewLeaderboardUseCase(repo repository.LeaderboardRepository) LeaderboardUseCase {
	return &leaderboardUseCase{repo: repo}
}

func (u *leaderboardUseCase) AddPoints(ctx context.Context, req dto.PointRequest) error {
	serviceID, err := u.repo.GetServiceIDByPartnerID(ctx, req.PartnerID)
	if err != nil {
		return err
	}

	if err := u.repo.SavePointHistory(ctx, req.PartnerID, req.Amount, req.Reason); err != nil {
		return err
	}

	return u.repo.IncrementScore(ctx, serviceID, req.PartnerID, float64(req.Amount))
}

func (u *leaderboardUseCase) GetLeaderboard(ctx context.Context, serviceID string, limit int64) (dto.LeaderboardResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	results, err := u.repo.GetTopRank(ctx, serviceID, limit)
	if err != nil {
		return dto.LeaderboardResponse{}, err
	}

	entries := make([]dto.LeaderboardEntry, 0, len(results))
	for i, z := range results {
		entries = append(entries, dto.LeaderboardEntry{
			Rank:      int64(i + 1),
			PartnerID: z.Member.(string),
			Score:     z.Score,
		})
	}

	return dto.LeaderboardResponse{Entries: entries}, nil
}

func (u *leaderboardUseCase) GetPartnerRank(ctx context.Context, userID, serviceID string) (dto.LeaderboardEntry, error) {
	partnerID, err := u.repo.GetPartnerIDByUserAndService(ctx, userID, serviceID)
	if err != nil {
		return dto.LeaderboardEntry{}, err
	}

	rank, score, err := u.repo.GetUserRank(ctx, serviceID, partnerID)
	if err != nil {
		return dto.LeaderboardEntry{}, err
	}

	return dto.LeaderboardEntry{
		Rank:      rank,
		PartnerID: partnerID,
		Score:     score,
	}, nil
}
