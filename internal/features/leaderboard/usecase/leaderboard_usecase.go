package usecase

import (
	"context"

	"fmt"
	"log/slog"

	leaderboardDTO "github.com/taqiyyaghazi/ecosystem-engine/internal/features/leaderboard/dto"
	leaderboardRepository "github.com/taqiyyaghazi/ecosystem-engine/internal/features/leaderboard/repository"
	notificationDTO "github.com/taqiyyaghazi/ecosystem-engine/internal/features/notifications/dto"
	notificationUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/notifications/usecase"
)

type LeaderboardUseCase interface {
	AddPoints(ctx context.Context, req leaderboardDTO.PointRequest) error
	GetLeaderboard(ctx context.Context, serviceID string, limit int64) (leaderboardDTO.LeaderboardResponse, error)
	GetPartnerRank(ctx context.Context, userID, serviceID string) (leaderboardDTO.LeaderboardEntry, error)
}

type leaderboardUseCase struct {
	leaderboardRepository leaderboardRepository.LeaderboardRepository
	notificationUseCase   notificationUsecase.NotificationUsecase
}

func NewLeaderboardUseCase(leaderboardRepository leaderboardRepository.LeaderboardRepository, notificationUseCase notificationUsecase.NotificationUsecase) LeaderboardUseCase {
	return &leaderboardUseCase{
		leaderboardRepository: leaderboardRepository,
		notificationUseCase:   notificationUseCase,
	}
}

func (u *leaderboardUseCase) AddPoints(ctx context.Context, req leaderboardDTO.PointRequest) error {
	serviceID, err := u.leaderboardRepository.GetServiceIDByPartnerID(ctx, req.PartnerID)
	if err != nil {
		return err
	}

	if err := u.leaderboardRepository.SavePointHistory(ctx, req.PartnerID, req.Amount, req.Reason); err != nil {
		return err
	}

	if err := u.leaderboardRepository.IncrementScore(ctx, serviceID, req.PartnerID, float64(req.Amount)); err != nil {
		return err
	}

	// Publish point update notification as fire-and-forget
	channel := "notifications:user:" + req.PartnerID
	go func() {
		msg := notificationDTO.NotificationMessage{
			RecipientID: req.PartnerID,
			Title:       "Points Added!",
			Message:     fmt.Sprintf("You have received %d points for: %s", req.Amount, req.Reason),
		}
		if err := u.notificationUseCase.Send(context.Background(), channel, msg); err != nil {
			slog.Error("failed to publish point update notification", "error", err, "partner_id", req.PartnerID)
		}
	}()

	return nil
}

func (u *leaderboardUseCase) GetLeaderboard(ctx context.Context, serviceID string, limit int64) (leaderboardDTO.LeaderboardResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	results, err := u.leaderboardRepository.GetTopRank(ctx, serviceID, limit)
	if err != nil {
		return leaderboardDTO.LeaderboardResponse{}, err
	}

	entries := make([]leaderboardDTO.LeaderboardEntry, 0, len(results))
	for i, z := range results {
		entries = append(entries, leaderboardDTO.LeaderboardEntry{
			Rank:      int64(i + 1),
			PartnerID: z.Member.(string),
			Score:     z.Score,
		})
	}

	return leaderboardDTO.LeaderboardResponse{Entries: entries}, nil
}

func (u *leaderboardUseCase) GetPartnerRank(ctx context.Context, userID, serviceID string) (leaderboardDTO.LeaderboardEntry, error) {
	partnerID, err := u.leaderboardRepository.GetPartnerIDByUserAndService(ctx, userID, serviceID)
	if err != nil {
		return leaderboardDTO.LeaderboardEntry{}, err
	}

	rank, score, err := u.leaderboardRepository.GetUserRank(ctx, serviceID, partnerID)
	if err != nil {
		return leaderboardDTO.LeaderboardEntry{}, err
	}

	return leaderboardDTO.LeaderboardEntry{
		Rank:      rank,
		PartnerID: partnerID,
		Score:     score,
	}, nil
}
