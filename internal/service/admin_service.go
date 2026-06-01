package service

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/repository"
	"github.com/jmoiron/sqlx"
	"github.com/xuri/excelize/v2"
)

type AdminService interface {
	GetAllBundles(ctx context.Context) ([]model.Bundle, error)
	GetBundleByPublicID(ctx context.Context, publicID string) (*model.Bundle, error)
	UploadBundle(ctx context.Context, fileReader io.Reader, mapelID int, namaBundle string, waktuMenit int, userID int64) error
	ExportBundle(ctx context.Context, bundleID int64) ([]byte, error)
	UpdateBundleWithExcel(ctx context.Context, bundleID int64, fileReader io.Reader) error
	GetSubmissions(ctx context.Context, status string) ([]map[string]interface{}, error)
	GetSubmissionDetail(ctx context.Context, historyID int64) (map[string]interface{}, error)
	GradeSubmission(ctx context.Context, historyID int64, req model.GradeRequest) error
}

type adminService struct {
	db          *sqlx.DB
	bundleRepo  repository.BundleRepository
	soalRepo    repository.SoalRepository
	historyRepo repository.HistoryRepository
	userRepo    repository.UserRepository
	jenjangRepo repository.JenjangRepository
	mapelRepo   repository.MapelRepository
}

func NewAdminService(
	db *sqlx.DB,
	bundleRepo repository.BundleRepository,
	soalRepo repository.SoalRepository,
	historyRepo repository.HistoryRepository,
	userRepo repository.UserRepository,
	jenjangRepo repository.JenjangRepository,
	mapelRepo repository.MapelRepository,
) AdminService {
	return &adminService{
		db:          db,
		bundleRepo:  bundleRepo,
		soalRepo:    soalRepo,
		historyRepo: historyRepo,
		userRepo:    userRepo,
		jenjangRepo: jenjangRepo,
		mapelRepo:   mapelRepo,
	}
}

func (s *adminService) GetAllBundles(ctx context.Context) ([]model.Bundle, error) {
	return s.bundleRepo.FindAll(ctx)
}

func (s *adminService) GetBundleByPublicID(ctx context.Context, publicID string) (*model.Bundle, error) {
	return s.bundleRepo.FindByPublicID(ctx, publicID)
}

func (s *adminService) UploadBundle(ctx context.Context, fileReader io.Reader, dummyMapelID int, dummyNamaBundle string, dummyWaktuMenit int, userID int64) error {
	f, err := excelize.OpenReader(fileReader)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return err
	}

	if len(rows) < 2 {
		return fmt.Errorf("file excel kosong atau tidak valid")
	}

	// Call pre-validation helper
	if err := validateUploadRows(rows); err != nil {
		return err
	}

	// Read first data row to get Jenjang, Mapel, Bundle Info
	firstDataRow := rows[1]
	if len(firstDataRow) < 4 {
		return fmt.Errorf("baris pertama data tidak lengkap")
	}

	jenjangNama := firstDataRow[0]
	mapelNama := firstDataRow[1]
	bundleNama := firstDataRow[2]
	waktuMenit, _ := strconv.Atoi(firstDataRow[3])

	// 1. Get Jenjang
	jenjang, err := s.jenjangRepo.FindByNama(ctx, jenjangNama)
	if err != nil {
		return fmt.Errorf("jenjang '%s' tidak ditemukan di database", jenjangNama)
	}

	// 2. Get Mapel
	mapel, err := s.mapelRepo.FindByNamaAndJenjang(ctx, mapelNama, jenjang.ID)
	if err != nil {
		return fmt.Errorf("mapel '%s' tidak ditemukan untuk jenjang '%s'", mapelNama, jenjangNama)
	}

	// Start Transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txCtx := repository.InjectTx(ctx, tx)

	// 3. Create Bundle
	bundle := &model.Bundle{
		MapelID:    mapel.ID,
		NamaBundle: bundleNama,
		WaktuMenit: waktuMenit,
		IsActive:   true,
		CreatedBy:  &userID,
	}

	err = s.bundleRepo.Create(txCtx, bundle)
	if err != nil {
		return err
	}

	var soalList []model.Soal
	for i, row := range rows {
		if i == 0 { // Skip header
			continue
		}
		if len(row) < 9 { // Updated index for soal data
			continue
		}

		tipe := model.SoalPG
		if strings.ToLower(row[4]) == "isian_singkat" {
			tipe = model.SoalIsianSingkat
		}

		bobot, _ := strconv.Atoi(row[8])
		if bobot == 0 {
			bobot = 1
		}

		pilihanJSON := model.PilihanJawabanList{}
		if tipe == model.SoalPG && len(row) >= 10 {
			opsi := []string{"A", "B", "C", "D", "E"}
			for j, teks := range row[9:] {
				if j < len(opsi) && teks != "" {
					pilihanJSON = append(pilihanJSON, model.PilihanJawaban{Opsi: opsi[j], Teks: teks})
				}
			}
		}

		soalList = append(soalList, model.Soal{
			BundleID:       bundle.ID,
			TipeSoal:       tipe,
			TeksSoal:       row[5],
			KunciJawaban:   row[6],
			Pembahasan:     row[7],
			BobotNilai:     bobot,
			PilihanJawaban: pilihanJSON,
		})
	}

	err = s.soalRepo.BulkCreate(txCtx, soalList)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *adminService) ExportBundle(ctx context.Context, bundleID int64) ([]byte, error) {
	soalList, err := s.soalRepo.FindByBundleID(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)

	f.SetCellValue(sheet, "A1", "ID Soal")
	f.SetCellValue(sheet, "B1", "Tipe Soal")
	f.SetCellValue(sheet, "C1", "Teks Soal")
	f.SetCellValue(sheet, "D1", "Kunci Jawaban")
	f.SetCellValue(sheet, "E1", "Pembahasan")
	f.SetCellValue(sheet, "F1", "Bobot Nilai")
	f.SetCellValue(sheet, "G1", "Pilihan A")
	f.SetCellValue(sheet, "H1", "Pilihan B")
	f.SetCellValue(sheet, "I1", "Pilihan C")
	f.SetCellValue(sheet, "J1", "Pilihan D")
	f.SetCellValue(sheet, "K1", "Pilihan E")

	for i, s := range soalList {
		rowIdx := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIdx), s.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIdx), s.TipeSoal)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowIdx), s.TeksSoal)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowIdx), s.KunciJawaban)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", rowIdx), s.Pembahasan)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", rowIdx), s.BobotNilai)

		if s.TipeSoal == model.SoalPG {
			for j, p := range s.PilihanJawaban {
				col := string(rune('G' + j))
				f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, rowIdx), p.Teks)
			}
		}
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (s *adminService) UpdateBundleWithExcel(ctx context.Context, bundleID int64, fileReader io.Reader) error {
	f, err := excelize.OpenReader(fileReader)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return err
	}

	// Call pre-validation helper
	if err := validateUpdateRows(rows); err != nil {
		return err
	}

	// Start Transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txCtx := repository.InjectTx(ctx, tx)

	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 6 {
			continue
		}

		tipe := model.SoalPG
		if strings.ToLower(row[1]) == "isian_singkat" {
			tipe = model.SoalIsianSingkat
		}

		bobot, _ := strconv.Atoi(row[5])
		if bobot == 0 {
			bobot = 1
		}

		pilihanJSON := model.PilihanJawabanList{}
		if tipe == model.SoalPG && len(row) >= 7 {
			opsi := []string{"A", "B", "C", "D", "E"}
			for j, teks := range row[6:] {
				if j < len(opsi) && teks != "" {
					pilihanJSON = append(pilihanJSON, model.PilihanJawaban{Opsi: opsi[j], Teks: teks})
				}
			}
		}

		soal := &model.Soal{
			BundleID:       bundleID,
			TipeSoal:       tipe,
			TeksSoal:       row[2],
			KunciJawaban:   row[3],
			Pembahasan:     row[4],
			BobotNilai:     bobot,
			PilihanJawaban: pilihanJSON,
		}

		idSoalStr := row[0]
		if idSoalStr != "" {
			idSoal, err := strconv.ParseInt(idSoalStr, 10, 64)
			if err != nil {
				return fmt.Errorf("id soal tidak valid pada baris %d: %v", i+1, err)
			}
			soal.ID = idSoal
			err = s.soalRepo.Update(txCtx, soal)
			if err != nil {
				return fmt.Errorf("gagal mengupdate soal pada baris %d: %v", i+1, err)
			}
		} else {
			err = s.soalRepo.BulkCreate(txCtx, []model.Soal{*soal})
			if err != nil {
				return fmt.Errorf("gagal menyisipkan soal baru pada baris %d: %v", i+1, err)
			}
		}
	}

	return tx.Commit()
}

func (s *adminService) GetSubmissions(ctx context.Context, status string) ([]map[string]interface{}, error) {
	st := model.StatusMenungguKoreksi
	if status != "" {
		st = model.StatusUjian(status)
	}

	histories, err := s.historyRepo.FindByStatus(ctx, st)
	if err != nil {
		return nil, err
	}

	var res []map[string]interface{}
	for _, h := range histories {
		username := "unknown"
		user, err := s.userRepo.FindByID(ctx, h.UserID)
		if err == nil && user != nil {
			username = user.Username
		}

		namaBundle := "unknown"
		bundle, err := s.bundleRepo.FindByID(ctx, h.BundleID)
		if err == nil && bundle != nil {
			namaBundle = bundle.NamaBundle
		}

		res = append(res, map[string]interface{}{
			"history_id":     h.ID,
			"username":       username,
			"nama_bundle":    namaBundle,
			"status":         h.Status,
			"tanggal_submit": h.WaktuSelesai,
		})
	}

	return res, nil
}

func (s *adminService) GetSubmissionDetail(ctx context.Context, historyID int64) (map[string]interface{}, error) {
	history, err := s.historyRepo.FindByID(ctx, historyID)
	if err != nil {
		return nil, err
	}

	username := "unknown"
	user, err := s.userRepo.FindByID(ctx, history.UserID)
	if err == nil && user != nil {
		username = user.Username
	}

	soalList, err := s.soalRepo.FindByBundleID(ctx, history.BundleID)
	if err != nil {
		return nil, err
	}
	soalMap := make(map[int64]model.Soal)
	for _, sl := range soalList {
		soalMap[sl.ID] = sl
	}

	var detailJawaban []map[string]interface{}
	for _, dj := range history.DetailJawaban {
		soal, ok := soalMap[dj.SoalID]
		teksSoal := "Soal tidak ditemukan"
		if ok {
			teksSoal = soal.TeksSoal
		}
		detailJawaban = append(detailJawaban, map[string]interface{}{
			"soal_id":         dj.SoalID,
			"teks_soal":       teksSoal,
			"jawaban_peserta": dj.JawabanPeserta,
			"skor_didapat":    dj.SkorDidapat,
			"is_dinilai":      dj.IsDinilai,
		})
	}

	return map[string]interface{}{
		"history_id":     history.ID,
		"username":       username,
		"detail_jawaban": detailJawaban,
	}, nil
}

func (s *adminService) GradeSubmission(ctx context.Context, historyID int64, req model.GradeRequest) error {
	history, err := s.historyRepo.FindByID(ctx, historyID)
	if err != nil {
		return err
	}

	gradeMap := make(map[int64]float64)
	for _, p := range req.PenilaianManual {
		gradeMap[p.SoalID] = p.SkorDiberikan
	}

	var totalSkor float64
	for i, dj := range history.DetailJawaban {
		if skor, ok := gradeMap[dj.SoalID]; ok {
			history.DetailJawaban[i].SkorDidapat = skor
			history.DetailJawaban[i].IsDinilai = true
		}
		totalSkor += history.DetailJawaban[i].SkorDidapat
	}

	history.SkorAkhir = totalSkor
	history.Status = model.StatusSelesai
	return s.historyRepo.Update(ctx, history)
}

func validateUploadRows(rows [][]string) error {
	for i, row := range rows {
		if i == 0 { // Skip header
			continue
		}
		
		// Check if row is completely empty
		isEmpty := true
		for _, val := range row {
			if strings.TrimSpace(val) != "" {
				isEmpty = false
				break
			}
		}
		if isEmpty {
			continue
		}

		if len(row) < 9 {
			return fmt.Errorf("baris %d: jumlah kolom kurang dari 9 (terdapat %d kolom)", i+1, len(row))
		}

		tipeStr := strings.ToLower(strings.TrimSpace(row[4]))
		if tipeStr != "pilihan_ganda" && tipeStr != "isian_singkat" {
			return fmt.Errorf("baris %d: tipe soal '%s' tidak valid (harus 'pilihan_ganda' atau 'isian_singkat')", i+1, row[4])
		}

		if strings.TrimSpace(row[5]) == "" {
			return fmt.Errorf("baris %d: teks soal tidak boleh kosong", i+1)
		}

		kunci := strings.ToUpper(strings.TrimSpace(row[6]))
		if kunci == "" {
			return fmt.Errorf("baris %d: kunci jawaban tidak boleh kosong", i+1)
		}

		if row[8] != "" {
			if _, err := strconv.Atoi(row[8]); err != nil {
				return fmt.Errorf("baris %d: bobot nilai '%s' tidak valid (harus angka)", i+1, row[8])
			}
		}

		if tipeStr == "pilihan_ganda" {
			// Must have choices columns
			if len(row) < 11 {
				return fmt.Errorf("baris %d: soal pilihan ganda harus memiliki minimal pilihan A dan B", i+1)
			}
			pilihanA := strings.TrimSpace(row[9])
			pilihanB := strings.TrimSpace(row[10])
			if pilihanA == "" || pilihanB == "" {
				return fmt.Errorf("baris %d: soal pilihan ganda harus memiliki minimal pilihan A dan B", i+1)
			}

			// Validate key matches one of options A..E
			validKeys := []string{"A", "B", "C", "D", "E"}
			isValidKey := false
			for j, k := range validKeys {
				colIdx := 9 + j
				if colIdx < len(row) && strings.TrimSpace(row[colIdx]) != "" {
					if kunci == k {
						isValidKey = true
						break
					}
				}
			}
			if !isValidKey {
				return fmt.Errorf("baris %d: kunci jawaban '%s' tidak cocok dengan pilihan jawaban yang tersedia", i+1, row[6])
			}
		}
	}
	return nil
}

func validateUpdateRows(rows [][]string) error {
	for i, row := range rows {
		if i == 0 {
			continue
		}
		
		isEmpty := true
		for _, val := range row {
			if strings.TrimSpace(val) != "" {
				isEmpty = false
				break
			}
		}
		if isEmpty {
			continue
		}

		if len(row) < 6 {
			return fmt.Errorf("baris %d: jumlah kolom kurang dari 6 (terdapat %d kolom)", i+1, len(row))
		}

		tipeStr := strings.ToLower(strings.TrimSpace(row[1]))
		if tipeStr != "pilihan_ganda" && tipeStr != "isian_singkat" {
			return fmt.Errorf("baris %d: tipe soal '%s' tidak valid (harus 'pilihan_ganda' atau 'isian_singkat')", i+1, row[1])
		}

		if strings.TrimSpace(row[2]) == "" {
			return fmt.Errorf("baris %d: teks soal tidak boleh kosong", i+1)
		}

		kunci := strings.ToUpper(strings.TrimSpace(row[3]))
		if kunci == "" {
			return fmt.Errorf("baris %d: kunci jawaban tidak boleh kosong", i+1)
		}

		if row[5] != "" {
			if _, err := strconv.Atoi(row[5]); err != nil {
				return fmt.Errorf("baris %d: bobot nilai '%s' tidak valid (harus angka)", i+1, row[5])
			}
		}

		if tipeStr == "pilihan_ganda" {
			if len(row) < 8 {
				return fmt.Errorf("baris %d: soal pilihan ganda harus memiliki minimal pilihan A dan B", i+1)
			}
			pilihanA := strings.TrimSpace(row[6])
			pilihanB := strings.TrimSpace(row[7])
			if pilihanA == "" || pilihanB == "" {
				return fmt.Errorf("baris %d: soal pilihan ganda harus memiliki minimal pilihan A dan B", i+1)
			}

			validKeys := []string{"A", "B", "C", "D", "E"}
			isValidKey := false
			for j, k := range validKeys {
				colIdx := 6 + j
				if colIdx < len(row) && strings.TrimSpace(row[colIdx]) != "" {
					if kunci == k {
						isValidKey = true
						break
					}
				}
			}
			if !isValidKey {
				return fmt.Errorf("baris %d: kunci jawaban '%s' tidak cocok dengan pilihan jawaban yang tersedia", i+1, row[3])
			}
		}
	}
	return nil
}
