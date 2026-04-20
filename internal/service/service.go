package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"ams-ai/internal/config"
	"ams-ai/internal/domain"
	"ams-ai/internal/platform/postgres"
	"ams-ai/internal/platform/storage"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	store     *postgres.Store
	objects   storage.ObjectStorage
	cfg       config.Config
	now       func() time.Time
	maxUpload int64
}

type AuthToken struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expiresAt"`
	User      domain.User `json:"user"`
}

type UploadInput struct {
	AssetID     int64
	Title       string
	Type        string
	Notes       string
	FileName    string
	ContentType string
	SizeBytes   int64
	Reader      io.Reader
	User        domain.User
}

func New(store *postgres.Store, objects storage.ObjectStorage, cfg config.Config) *Service {
	return &Service{
		store:     store,
		objects:   objects,
		cfg:       cfg,
		now:       time.Now,
		maxUpload: cfg.Storage.MaxUploadBytes,
	}
}

func (s *Service) Login(ctx context.Context, email, password string) (AuthToken, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return AuthToken{}, domain.ErrUnauthorized
		}
		return AuthToken{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return AuthToken{}, domain.ErrUnauthorized
	}
	expiresAt := s.now().Add(s.cfg.TokenTTL)
	token, err := s.signToken(user.ID, user.Role, expiresAt)
	if err != nil {
		return AuthToken{}, err
	}
	return AuthToken{Token: token, ExpiresAt: expiresAt, User: user}, nil
}

func (s *Service) UserFromToken(ctx context.Context, token string) (domain.User, error) {
	userID, role, expiresAt, err := s.verifyToken(token)
	if err != nil {
		return domain.User{}, domain.ErrUnauthorized
	}
	if s.now().After(expiresAt) {
		return domain.User{}, domain.ErrUnauthorized
	}
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}
	if user.Role != role {
		return domain.User{}, domain.ErrUnauthorized
	}
	return user, nil
}

func (s *Service) ListUsers(ctx context.Context, user domain.User) ([]domain.User, error) {
	if user.Role != domain.RoleAdmin {
		return []domain.User{user}, nil
	}
	return s.store.ListUsers(ctx)
}

func (s *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.store.ListCategories(ctx)
}

func (s *Service) CreateCategory(ctx context.Context, user domain.User, name, description string) (domain.Category, error) {
	if user.Role != domain.RoleAdmin {
		return domain.Category{}, domain.ErrForbidden
	}
	if strings.TrimSpace(name) == "" {
		return domain.Category{}, fmt.Errorf("%w: category name is required", domain.ErrInvalid)
	}
	return s.store.CreateCategory(ctx, name, description)
}

func (s *Service) UpdateCategory(ctx context.Context, user domain.User, id int64, name, description string) (domain.Category, error) {
	if user.Role != domain.RoleAdmin {
		return domain.Category{}, domain.ErrForbidden
	}
	if strings.TrimSpace(name) == "" {
		return domain.Category{}, fmt.Errorf("%w: category name is required", domain.ErrInvalid)
	}
	return s.store.UpdateCategory(ctx, id, name, description)
}

func (s *Service) CreateAsset(ctx context.Context, user domain.User, a domain.Asset) (domain.Asset, error) {
	if err := validateAsset(&a); err != nil {
		return domain.Asset{}, err
	}
	a.CreatedBy = user.ID
	a.UpdatedBy = &user.ID
	if user.Role != domain.RoleAdmin {
		a.AssignedUserID = &user.ID
	}
	return s.store.CreateAsset(ctx, a)
}

func (s *Service) UpdateAsset(ctx context.Context, user domain.User, a domain.Asset) (domain.Asset, error) {
	existing, err := s.store.GetAsset(ctx, a.ID)
	if err != nil {
		return domain.Asset{}, err
	}
	if !domain.AssetAccessAllowed(user, existing) {
		return domain.Asset{}, domain.ErrForbidden
	}
	if err := validateAsset(&a); err != nil {
		return domain.Asset{}, err
	}
	if user.Role != domain.RoleAdmin {
		a.AssignedUserID = existing.AssignedUserID
		if a.AssignedUserID == nil {
			a.AssignedUserID = &user.ID
		}
	}
	a.UpdatedBy = &user.ID
	return s.store.UpdateAsset(ctx, a)
}

func (s *Service) ArchiveAsset(ctx context.Context, user domain.User, id int64) error {
	asset, err := s.store.GetAsset(ctx, id)
	if err != nil {
		return err
	}
	if !domain.AssetAccessAllowed(user, asset) {
		return domain.ErrForbidden
	}
	return s.store.ArchiveAsset(ctx, id, user.ID)
}

func (s *Service) RestoreAsset(ctx context.Context, user domain.User, id int64) error {
	asset, err := s.store.GetAsset(ctx, id)
	if err != nil {
		return err
	}
	if !domain.AssetAccessAllowed(user, asset) {
		return domain.ErrForbidden
	}
	return s.store.RestoreAsset(ctx, id, user.ID)
}

func (s *Service) GetAsset(ctx context.Context, user domain.User, id int64) (domain.Asset, error) {
	asset, err := s.store.GetAsset(ctx, id)
	if err != nil {
		return domain.Asset{}, err
	}
	if !domain.AssetAccessAllowed(user, asset) {
		return domain.Asset{}, domain.ErrForbidden
	}
	return asset, nil
}

func (s *Service) ListAssets(ctx context.Context, user domain.User, f domain.AssetFilters) ([]domain.Asset, error) {
	f.CurrentUserID = user.ID
	f.CurrentUserRole = user.Role
	f.ReminderWindowDay = s.cfg.ReminderWindowDays
	return s.store.ListAssets(ctx, f)
}

func (s *Service) UpsertVehicleProfile(ctx context.Context, user domain.User, p domain.VehicleProfile) (domain.VehicleProfile, error) {
	asset, err := s.GetAsset(ctx, user, p.AssetID)
	if err != nil {
		return domain.VehicleProfile{}, err
	}
	if asset.Type != domain.AssetTypeVehicle {
		return domain.VehicleProfile{}, fmt.Errorf("%w: vehicle profile requires a vehicle asset", domain.ErrInvalid)
	}
	if strings.TrimSpace(p.RegistrationNumber) == "" {
		return domain.VehicleProfile{}, fmt.Errorf("%w: registration number is required", domain.ErrInvalid)
	}
	return s.store.UpsertVehicleProfile(ctx, p)
}

func (s *Service) GetVehicleProfile(ctx context.Context, user domain.User, assetID int64) (domain.VehicleProfile, error) {
	asset, err := s.GetAsset(ctx, user, assetID)
	if err != nil {
		return domain.VehicleProfile{}, err
	}
	if asset.Type != domain.AssetTypeVehicle {
		return domain.VehicleProfile{}, fmt.Errorf("%w: asset is not a vehicle", domain.ErrInvalid)
	}
	return s.store.GetVehicleProfile(ctx, assetID)
}

func (s *Service) CreateInsuranceRecord(ctx context.Context, user domain.User, r domain.VehicleInsuranceRecord) (domain.VehicleInsuranceRecord, error) {
	if err := s.ensureVehicleAccess(ctx, user, r.AssetID); err != nil {
		return domain.VehicleInsuranceRecord{}, err
	}
	if r.ExpiryDate.IsZero() {
		return domain.VehicleInsuranceRecord{}, fmt.Errorf("%w: insurance expiry date is required", domain.ErrInvalid)
	}
	return s.store.CreateInsuranceRecord(ctx, r)
}

func (s *Service) ListInsuranceRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.VehicleInsuranceRecord, error) {
	if err := s.ensureVehicleAccess(ctx, user, assetID); err != nil {
		return nil, err
	}
	return s.store.ListInsuranceRecords(ctx, assetID)
}

func (s *Service) CreateLicenseRecord(ctx context.Context, user domain.User, r domain.VehicleLicenseRecord) (domain.VehicleLicenseRecord, error) {
	if err := s.ensureVehicleAccess(ctx, user, r.AssetID); err != nil {
		return domain.VehicleLicenseRecord{}, err
	}
	if r.ExpiryDate.IsZero() {
		return domain.VehicleLicenseRecord{}, fmt.Errorf("%w: license expiry date is required", domain.ErrInvalid)
	}
	return s.store.CreateLicenseRecord(ctx, r)
}

func (s *Service) ListLicenseRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.VehicleLicenseRecord, error) {
	if err := s.ensureVehicleAccess(ctx, user, assetID); err != nil {
		return nil, err
	}
	return s.store.ListLicenseRecords(ctx, assetID)
}

func (s *Service) CreateEmissionRecord(ctx context.Context, user domain.User, r domain.VehicleEmissionRecord) (domain.VehicleEmissionRecord, error) {
	if err := s.ensureVehicleAccess(ctx, user, r.AssetID); err != nil {
		return domain.VehicleEmissionRecord{}, err
	}
	if r.ExpiryDate.IsZero() {
		return domain.VehicleEmissionRecord{}, fmt.Errorf("%w: emission/inspection expiry date is required", domain.ErrInvalid)
	}
	return s.store.CreateEmissionRecord(ctx, r)
}

func (s *Service) ListEmissionRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.VehicleEmissionRecord, error) {
	if err := s.ensureVehicleAccess(ctx, user, assetID); err != nil {
		return nil, err
	}
	return s.store.ListEmissionRecords(ctx, assetID)
}

func (s *Service) CreateServiceRecord(ctx context.Context, user domain.User, r domain.ServiceRecord) (domain.ServiceRecord, error) {
	if _, err := s.GetAsset(ctx, user, r.AssetID); err != nil {
		return domain.ServiceRecord{}, err
	}
	if r.ServiceDate.IsZero() {
		return domain.ServiceRecord{}, fmt.Errorf("%w: service date is required", domain.ErrInvalid)
	}
	if r.ServiceType == "" {
		r.ServiceType = "service"
	}
	if r.ServiceType != "service" && r.ServiceType != "repair" {
		return domain.ServiceRecord{}, fmt.Errorf("%w: service type must be service or repair", domain.ErrInvalid)
	}
	r.CreatedBy = user.ID
	return s.store.CreateServiceRecord(ctx, r)
}

func (s *Service) ListServiceRecords(ctx context.Context, user domain.User, assetID int64) ([]domain.ServiceRecord, error) {
	if _, err := s.GetAsset(ctx, user, assetID); err != nil {
		return nil, err
	}
	return s.store.ListServiceRecords(ctx, assetID)
}

func (s *Service) CreateFuelLog(ctx context.Context, user domain.User, r domain.FuelLog) (domain.FuelLog, error) {
	if err := s.ensureVehicleAccess(ctx, user, r.AssetID); err != nil {
		return domain.FuelLog{}, err
	}
	if r.FuelDate.IsZero() {
		return domain.FuelLog{}, fmt.Errorf("%w: fuel date is required", domain.ErrInvalid)
	}
	if r.Quantity <= 0 || r.Cost < 0 {
		return domain.FuelLog{}, fmt.Errorf("%w: fuel quantity and cost must be valid", domain.ErrInvalid)
	}
	r.CreatedBy = user.ID
	return s.store.CreateFuelLog(ctx, r)
}

func (s *Service) ListFuelLogs(ctx context.Context, user domain.User, assetID int64) ([]domain.FuelLog, error) {
	if err := s.ensureVehicleAccess(ctx, user, assetID); err != nil {
		return nil, err
	}
	return s.store.ListFuelLogs(ctx, assetID)
}

func (s *Service) UploadDocument(ctx context.Context, in UploadInput) (domain.AssetDocument, error) {
	if _, err := s.GetAsset(ctx, in.User, in.AssetID); err != nil {
		return domain.AssetDocument{}, err
	}
	if strings.TrimSpace(in.Title) == "" {
		in.Title = in.FileName
	}
	if !validDocumentType(in.Type) {
		return domain.AssetDocument{}, fmt.Errorf("%w: unsupported document type", domain.ErrInvalid)
	}
	if !validContentType(in.ContentType) {
		return domain.AssetDocument{}, fmt.Errorf("%w: unsupported file type; only JPG, PNG, and PDF files are supported", domain.ErrInvalid)
	}
	if in.SizeBytes <= 0 || in.SizeBytes > s.maxUpload {
		return domain.AssetDocument{}, fmt.Errorf("%w: file size must be greater than 0 bytes and no larger than %s", domain.ErrInvalid, formatByteLimit(s.maxUpload))
	}
	key := fmt.Sprintf("assets/%d/documents/%d/%s", in.AssetID, s.now().UnixNano(), safeFileName(in.FileName))
	if err := s.objects.Put(ctx, key, in.Reader, in.SizeBytes, in.ContentType); err != nil {
		return domain.AssetDocument{}, err
	}
	doc, err := s.store.CreateDocument(ctx, domain.AssetDocument{
		AssetID:     in.AssetID,
		Title:       strings.TrimSpace(in.Title),
		Type:        in.Type,
		Notes:       strings.TrimSpace(in.Notes),
		FileName:    safeFileName(in.FileName),
		ContentType: in.ContentType,
		SizeBytes:   in.SizeBytes,
		ObjectKey:   key,
		UploadedBy:  in.User.ID,
	})
	if err != nil {
		_ = s.objects.Delete(ctx, key)
		return domain.AssetDocument{}, err
	}
	return doc, nil
}

func (s *Service) ListDocuments(ctx context.Context, user domain.User, assetID int64) ([]domain.AssetDocument, error) {
	if _, err := s.GetAsset(ctx, user, assetID); err != nil {
		return nil, err
	}
	return s.store.ListDocuments(ctx, assetID)
}

func (s *Service) DownloadDocument(ctx context.Context, user domain.User, id int64) (domain.AssetDocument, io.ReadCloser, storage.ObjectInfo, error) {
	doc, err := s.store.GetDocument(ctx, id)
	if err != nil {
		return domain.AssetDocument{}, nil, storage.ObjectInfo{}, err
	}
	if _, err := s.GetAsset(ctx, user, doc.AssetID); err != nil {
		return domain.AssetDocument{}, nil, storage.ObjectInfo{}, err
	}
	reader, info, err := s.objects.Get(ctx, doc.ObjectKey)
	if err != nil {
		return domain.AssetDocument{}, nil, storage.ObjectInfo{}, err
	}
	return doc, reader, info, nil
}

func (s *Service) DeleteDocument(ctx context.Context, user domain.User, id int64) error {
	doc, err := s.store.GetDocument(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.GetAsset(ctx, user, doc.AssetID); err != nil {
		return err
	}
	if err := s.store.DeleteDocument(ctx, id); err != nil {
		return err
	}
	return s.objects.Delete(ctx, doc.ObjectKey)
}

func (s *Service) Dashboard(ctx context.Context, user domain.User) (domain.Dashboard, error) {
	if err := s.store.RegenerateReminders(ctx, s.cfg.ReminderWindowDays); err != nil {
		return domain.Dashboard{}, err
	}
	return s.store.Dashboard(ctx, user, s.cfg.ReminderWindowDays)
}

func (s *Service) ListReminders(ctx context.Context, user domain.User) ([]domain.Reminder, error) {
	if err := s.store.RegenerateReminders(ctx, s.cfg.ReminderWindowDays); err != nil {
		return nil, err
	}
	return s.store.ListReminders(ctx, user, 100)
}

func (s *Service) RegenerateReminders(ctx context.Context, user domain.User) error {
	if user.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	return s.store.RegenerateReminders(ctx, s.cfg.ReminderWindowDays)
}

func (s *Service) ensureVehicleAccess(ctx context.Context, user domain.User, assetID int64) error {
	asset, err := s.GetAsset(ctx, user, assetID)
	if err != nil {
		return err
	}
	if asset.Type != domain.AssetTypeVehicle {
		return fmt.Errorf("%w: operation requires a vehicle asset", domain.ErrInvalid)
	}
	return nil
}

func validateAsset(a *domain.Asset) error {
	a.Type = strings.TrimSpace(strings.ToLower(a.Type))
	if a.Type == "" {
		a.Type = domain.AssetTypeGeneral
	}
	if !domain.IsValidAssetType(a.Type) {
		return fmt.Errorf("%w: asset type must be general or vehicle", domain.ErrInvalid)
	}
	a.Status = domain.NormalizeStatus(a.Status)
	if !domain.IsValidStatus(a.Status) {
		return fmt.Errorf("%w: invalid asset status", domain.ErrInvalid)
	}
	a.Name = strings.TrimSpace(a.Name)
	if a.Name == "" {
		return fmt.Errorf("%w: asset name is required", domain.ErrInvalid)
	}
	if a.CategoryID <= 0 {
		return fmt.Errorf("%w: category is required", domain.ErrInvalid)
	}
	a.Brand = strings.TrimSpace(a.Brand)
	a.Model = strings.TrimSpace(a.Model)
	a.SerialNumber = strings.TrimSpace(a.SerialNumber)
	a.Condition = strings.TrimSpace(a.Condition)
	a.Location = strings.TrimSpace(a.Location)
	a.AssignedTo = strings.TrimSpace(a.AssignedTo)
	a.Notes = strings.TrimSpace(a.Notes)
	a.WarrantyNotes = strings.TrimSpace(a.WarrantyNotes)
	if a.PurchasePrice != nil && *a.PurchasePrice < 0 {
		return fmt.Errorf("%w: purchase price cannot be negative", domain.ErrInvalid)
	}
	if a.WarrantyStartDate != nil && a.WarrantyExpiryDate != nil && a.WarrantyExpiryDate.Before(*a.WarrantyStartDate) {
		return fmt.Errorf("%w: warranty expiry cannot be before warranty start", domain.ErrInvalid)
	}
	return nil
}

func validDocumentType(t string) bool {
	switch t {
	case "bill_invoice", "warranty", "insurance", "license_registration", "service_receipt", "manual", "other":
		return true
	default:
		return false
	}
}

func validContentType(t string) bool {
	if mediaType, _, err := mime.ParseMediaType(t); err == nil {
		t = mediaType
	}
	switch t {
	case "image/jpeg", "image/png", "application/pdf":
		return true
	default:
		return false
	}
}

func safeFileName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "." || name == "/" || name == "" {
		return "upload"
	}
	var b strings.Builder
	for _, r := range name {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
		case r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "._-")
	if out == "" {
		return "upload"
	}
	return out
}

func formatByteLimit(limit int64) string {
	if limit%(1024*1024) == 0 {
		return fmt.Sprintf("%d MB", limit/(1024*1024))
	}
	return fmt.Sprintf("%d bytes", limit)
}

func (s *Service) signToken(userID int64, role string, expiresAt time.Time) (string, error) {
	payload := fmt.Sprintf("%d|%s|%d", userID, role, expiresAt.Unix())
	mac := hmac.New(sha256.New, []byte(s.cfg.AuthSecret))
	if _, err := mac.Write([]byte(payload)); err != nil {
		return "", err
	}
	sig := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (s *Service) verifyToken(token string) (int64, string, time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.AuthSecret))
	_, _ = mac.Write(payloadBytes)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	fields := strings.Split(string(payloadBytes), "|")
	if len(fields) != 3 {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	userID, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	role := fields[1]
	if !domain.IsValidRole(role) {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	exp, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	return userID, role, time.Unix(exp, 0), nil
}

func BearerToken(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
