package service

import (
	"context"
	"errors"
	"time"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/repository"
)

type ParticipantService interface {
	GetJenjang(ctx context.Context) ([]model.Jenjang, error)
	GetMapelByJenjang(ctx context.Context, jenjangID int) ([]model.Mapel, error)
	GetBundlesByMapel(ctx context.Context, mapelID int) ([]model.Bundle, error)
	GetBundleByPublicID(ctx context.Context, publicID string) (*model.Bundle, error)
	GetSoalByBundle(ctx context.Context, userID int64, bundleID int64) ([]model.SoalPublic, error)
	SubmitUjian(ctx context.Context, userID int64, bundleID int64, req model.SubmitRequest) (*model.WebResponse, error)
	GetHistory(ctx context.Context, userID int64) ([]model.HistoryResponse, error)
	GetReview(ctx context.Context, userID int64, bundleID int64) ([]model.ReviewResponse, error)
}

type participantService struct {
	jenjangRepo repository.JenjangRepository
	mapelRepo   repository.MapelRepository
	bundleRepo  repository.BundleRepository
	soalRepo    repository.SoalRepository
	historyRepo repository.HistoryRepository
}

func NewParticipantService(
	jenjangRepo repository.JenjangRepository,
	mapelRepo repository.MapelRepository,
	bundleRepo repository.BundleRepository,
	soalRepo repository.SoalRepository,
	historyRepo repository.HistoryRepository,
) ParticipantService {
	return &participantService{
		jenjangRepo: jenjangRepo,
		mapelRepo:   mapelRepo,
		bundleRepo:  bundleRepo,
		soalRepo:    soalRepo,
		historyRepo: historyRepo,
	}
}

func (s *participantService) GetJenjang(ctx context.Context) ([]model.Jenjang, error) {
	return s.jenjangRepo.FindAll(ctx)
}

func (s *participantService) GetMapelByJenjang(ctx context.Context, jenjangID int) ([]model.Mapel, error) {
	return s.mapelRepo.FindByJenjangID(ctx, jenjangID)
}

func (s *participantService) GetBundlesByMapel(ctx context.Context, mapelID int) ([]model.Bundle, error) {
	return s.bundleRepo.FindByMapelID(ctx, mapelID, true)
}

func (s *participantService) GetBundleByPublicID(ctx context.Context, publicID string) (*model.Bundle, error) {
	return s.bundleRepo.FindByPublicID(ctx, publicID)
}

func (s *participantService) GetSoalByBundle(ctx context.Context, userID int64, bundleID int64) ([]model.SoalPublic, error) {
	soalList, err := s.soalRepo.FindByBundleID(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	history, err := s.historyRepo.FindOngoingHistoryByBundleID(ctx, userID, bundleID)
	if err != nil || history == nil {
		newHistory := &model.HistoryUjian{
			UserID:   userID,
			BundleID: bundleID,
			Status:   model.StatusBerlangsung,
		}
		err = s.historyRepo.Create(ctx, newHistory)
		if err != nil {
			// Fallback: If it failed (possibly due to concurrent duplicate insertion constraints),
			// try to search again to see if another concurrent thread succeeded.
			history, err = s.historyRepo.FindOngoingHistoryByBundleID(ctx, userID, bundleID)
			if err != nil || history == nil {
				return nil, errors.New("gagal memulai ujian baru: " + err.Error())
			}
		}
	}

	var publicSoal []model.SoalPublic
	for _, s := range soalList {
		publicSoal = append(publicSoal, model.SoalPublic{
			ID:             s.ID,
			TipeSoal:       s.TipeSoal,
			TeksSoal:       s.TeksSoal,
			PilihanJawaban: s.PilihanJawaban,
			BobotNilai:     s.BobotNilai,
		})
	}

	return publicSoal, nil
}

func (s *participantService) SubmitUjian(ctx context.Context, userID int64, bundleID int64, req model.SubmitRequest) (*model.WebResponse, error) {
	history, err := s.historyRepo.FindOngoingHistoryByBundleID(ctx, userID, bundleID)

	if err != nil || history == nil {
		return nil, errors.New("tidak ada ujian yang sedang berlangsung untuk bundle ini")
	}

	soalList, err := s.soalRepo.FindByBundleID(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	// Map user's answers by SoalID
	jawabanMap := make(map[int64]string)
	for _, j := range req.Jawaban {
		jawabanMap[j.SoalID] = j.JawabanPeserta
	}

	var totalSkor float64
	var detailJawaban []model.DetailJawaban
	hasIsianSingkat := false

	for _, soal := range soalList {
		jawabanPeserta, ok := jawabanMap[soal.ID]
		
		dj := model.DetailJawaban{
			SoalID:         soal.ID,
			JawabanPeserta: jawabanPeserta, // Will be empty string if not answered
			SkorDidapat:    0,
			IsDinilai:      false,
		}

		if soal.TipeSoal == model.SoalPG {
			dj.IsDinilai = true
			if ok && jawabanPeserta == soal.KunciJawaban {
				dj.SkorDidapat = float64(soal.BobotNilai)
				totalSkor += dj.SkorDidapat
			}
		} else {
			dj.SkorDidapat = 0
			dj.IsDinilai = false
			hasIsianSingkat = true
		}
		detailJawaban = append(detailJawaban, dj)
	}

	now := time.Now()
	history.WaktuSelesai = &now
	history.SkorAkhir = totalSkor
	history.DetailJawaban = detailJawaban
	if hasIsianSingkat {
		history.Status = model.StatusMenungguKoreksi
	} else {
		history.Status = model.StatusSelesai
	}

	err = s.historyRepo.Update(ctx, history)
	if err != nil {
		return nil, err
	}

	return &model.WebResponse{
		Message: "Jawaban berhasil disubmit",
		Data: map[string]interface{}{
			"status":         history.Status,
			"skor_sementara": totalSkor,
		},
	}, nil
}

func (s *participantService) GetHistory(ctx context.Context, userID int64) ([]model.HistoryResponse, error) {
	histories, err := s.historyRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var res []model.HistoryResponse
	for _, h := range histories {
		bundle, _ := s.bundleRepo.FindByID(ctx, h.BundleID)
		namaBundle := "Unknown"
		if bundle != nil {
			namaBundle = bundle.NamaBundle
		}
		res = append(res, model.HistoryResponse{
			HistoryID:  h.ID,
			NamaBundle: namaBundle,
			WaktuMulai: h.WaktuMulai,
			SkorAkhir:  h.SkorAkhir,
			Status:     h.Status,
		})
	}

	return res, nil
}

func (s *participantService) GetReview(ctx context.Context, userID int64, bundleID int64) ([]model.ReviewResponse, error) {
	// Temukan history ujian yang sudah selesai (bukan yang sedang berlangsung)
	histories, err := s.historyRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var history *model.HistoryUjian
	for _, h := range histories {
		if h.BundleID == bundleID && (h.Status == model.StatusSelesai || h.Status == model.StatusMenungguKoreksi) {
			history = &h
			break
		}
	}

	if history == nil {
		return nil, errors.New("not found: history ujian tidak ditemukan")
	}

	if history.Status == model.StatusMenungguKoreksi {
		return nil, errors.New("akses ditolak: status masih menunggu koreksi")
	}
// ...

	soalList, err := s.soalRepo.FindByBundleID(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	jawabanMap := make(map[int64]model.DetailJawaban)
	for _, j := range history.DetailJawaban {
		jawabanMap[j.SoalID] = j
	}

	var res []model.ReviewResponse
	for _, soal := range soalList {
		jawaban, ok := jawabanMap[soal.ID]
		
		isBenar := false
		if ok && soal.TipeSoal == model.SoalPG && jawaban.JawabanPeserta == soal.KunciJawaban {
			isBenar = true
		}

		res = append(res, model.ReviewResponse{
			ID:             soal.ID,
			TipeSoal:       soal.TipeSoal,
			TeksSoal:       soal.TeksSoal,
			PilihanJawaban: soal.PilihanJawaban,
			Pembahasan:     soal.Pembahasan,
			JawabanPeserta: func() string {
				if ok {
					return jawaban.JawabanPeserta
				}
				return ""
			}(),
			KunciJawaban: soal.KunciJawaban,
			IsBenar:      isBenar,
		})
	}

	return res, nil
}
