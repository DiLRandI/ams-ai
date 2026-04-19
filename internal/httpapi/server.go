package httpapi

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"ams-ai/internal/config"
	"ams-ai/internal/domain"
	"ams-ai/internal/service"
)

type Server struct {
	cfg     config.Config
	service *service.Service
	log     *slog.Logger
}

type contextKey string

const userKey contextKey = "user"

func New(cfg config.Config, svc *service.Service, log *slog.Logger) http.Handler {
	s := &Server{cfg: cfg, service: svc, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("POST /api/auth/login", s.login)

	mux.HandleFunc("GET /api/auth/me", s.requireAuth(s.me))
	mux.HandleFunc("GET /api/users", s.requireAuth(s.listUsers))
	mux.HandleFunc("GET /api/categories", s.requireAuth(s.listCategories))
	mux.HandleFunc("POST /api/categories", s.requireAuth(s.createCategory))
	mux.HandleFunc("PUT /api/categories/{id}", s.requireAuth(s.updateCategory))

	mux.HandleFunc("GET /api/assets", s.requireAuth(s.listAssets))
	mux.HandleFunc("POST /api/assets", s.requireAuth(s.createAsset))
	mux.HandleFunc("GET /api/assets/{id}", s.requireAuth(s.getAsset))
	mux.HandleFunc("PUT /api/assets/{id}", s.requireAuth(s.updateAsset))
	mux.HandleFunc("POST /api/assets/{id}/archive", s.requireAuth(s.archiveAsset))
	mux.HandleFunc("POST /api/assets/{id}/restore", s.requireAuth(s.restoreAsset))

	mux.HandleFunc("GET /api/assets/{id}/documents", s.requireAuth(s.listDocuments))
	mux.HandleFunc("POST /api/assets/{id}/documents", s.requireAuth(s.uploadDocument))
	mux.HandleFunc("GET /api/documents/{id}/download", s.requireAuth(s.downloadDocument))
	mux.HandleFunc("DELETE /api/documents/{id}", s.requireAuth(s.deleteDocument))

	mux.HandleFunc("GET /api/assets/{id}/vehicle", s.requireAuth(s.getVehicleProfile))
	mux.HandleFunc("PUT /api/assets/{id}/vehicle", s.requireAuth(s.upsertVehicleProfile))
	mux.HandleFunc("GET /api/assets/{id}/vehicle/insurance", s.requireAuth(s.listInsurance))
	mux.HandleFunc("POST /api/assets/{id}/vehicle/insurance", s.requireAuth(s.createInsurance))
	mux.HandleFunc("GET /api/assets/{id}/vehicle/licenses", s.requireAuth(s.listLicenses))
	mux.HandleFunc("POST /api/assets/{id}/vehicle/licenses", s.requireAuth(s.createLicense))
	mux.HandleFunc("GET /api/assets/{id}/vehicle/emissions", s.requireAuth(s.listEmissions))
	mux.HandleFunc("POST /api/assets/{id}/vehicle/emissions", s.requireAuth(s.createEmission))
	mux.HandleFunc("GET /api/assets/{id}/services", s.requireAuth(s.listServices))
	mux.HandleFunc("POST /api/assets/{id}/services", s.requireAuth(s.createService))
	mux.HandleFunc("GET /api/assets/{id}/fuel-logs", s.requireAuth(s.listFuelLogs))
	mux.HandleFunc("POST /api/assets/{id}/fuel-logs", s.requireAuth(s.createFuelLog))

	mux.HandleFunc("GET /api/dashboard", s.requireAuth(s.dashboard))
	mux.HandleFunc("GET /api/reminders", s.requireAuth(s.reminders))
	mux.HandleFunc("POST /api/reminders/regenerate", s.requireAuth(s.regenerateReminders))

	mux.HandleFunc("GET /api/reports/assets.csv", s.requireAuth(s.assetReport))
	mux.HandleFunc("GET /api/reports/warranties.csv", s.requireAuth(s.warrantyReport))
	mux.HandleFunc("GET /api/reports/vehicle-renewals.csv", s.requireAuth(s.vehicleRenewalReport))
	mux.HandleFunc("GET /api/reports/service-history.csv", s.requireAuth(s.serviceReport))
	mux.HandleFunc("GET /api/reports/fuel-logs.csv", s.requireAuth(s.fuelReport))

	return s.recover(s.cors(mux))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	token, err := s.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, token)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, currentUser(r))
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.service.ListUsers(r.Context(), currentUser(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (s *Server) listCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.service.ListCategories(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

func (s *Server) createCategory(w http.ResponseWriter, r *http.Request) {
	var req categoryRequest
	if !readJSON(w, r, &req) {
		return
	}
	c, err := s.service.CreateCategory(r.Context(), currentUser(r), req.Name, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) updateCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req categoryRequest
	if !readJSON(w, r, &req) {
		return
	}
	c, err := s.service.UpdateCategory(r.Context(), currentUser(r), id, req.Name, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) listAssets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var hasDocs *bool
	if raw := q.Get("hasDocuments"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err == nil {
			hasDocs = &v
		}
	}
	filters := domain.AssetFilters{
		Query:           q.Get("q"),
		Status:          normalizeOptionalStatus(q.Get("status")),
		Location:        q.Get("location"),
		WarrantyState:   q.Get("warrantyState"),
		HasDocuments:    hasDocs,
		IncludeArchived: q.Get("includeArchived") == "true",
	}
	filters.CategoryID, _ = strconv.ParseInt(q.Get("categoryId"), 10, 64)
	filters.AssignedUserID, _ = strconv.ParseInt(q.Get("assignedUserId"), 10, 64)
	assets, err := s.service.ListAssets(r.Context(), currentUser(r), filters)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, assets)
}

func (s *Server) createAsset(w http.ResponseWriter, r *http.Request) {
	var req assetRequest
	if !readJSON(w, r, &req) {
		return
	}
	asset, err := req.toAsset()
	if err != nil {
		writeError(w, err)
		return
	}
	created, err := s.service.CreateAsset(r.Context(), currentUser(r), asset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getAsset(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	asset, err := s.service.GetAsset(r.Context(), currentUser(r), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, asset)
}

func (s *Server) updateAsset(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req assetRequest
	if !readJSON(w, r, &req) {
		return
	}
	asset, err := req.toAsset()
	if err != nil {
		writeError(w, err)
		return
	}
	asset.ID = id
	updated, err := s.service.UpdateAsset(r.Context(), currentUser(r), asset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) archiveAsset(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := s.service.ArchiveAsset(r.Context(), currentUser(r), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) restoreAsset(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := s.service.RestoreAsset(r.Context(), currentUser(r), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listDocuments(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	docs, err := s.service.ListDocuments(r.Context(), currentUser(r), assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, docs)
}

func (s *Server) uploadDocument(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.Storage.MaxUploadBytes+1024*1024)
	if err := r.ParseMultipartForm(s.cfg.Storage.MaxUploadBytes); err != nil {
		writeError(w, fmt.Errorf("%w: invalid multipart upload", domain.ErrInvalid))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, fmt.Errorf("%w: file is required", domain.ErrInvalid))
		return
	}
	defer file.Close()
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(strings.ToLower(filepathExt(header.Filename)))
	}
	doc, err := s.service.UploadDocument(r.Context(), service.UploadInput{
		AssetID:     assetID,
		Title:       r.FormValue("title"),
		Type:        r.FormValue("type"),
		Notes:       r.FormValue("notes"),
		FileName:    header.Filename,
		ContentType: contentType,
		SizeBytes:   header.Size,
		Reader:      file,
		User:        currentUser(r),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, doc)
}

func (s *Server) downloadDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	doc, reader, info, err := s.service.DownloadDocument(r.Context(), currentUser(r), id)
	if err != nil {
		writeError(w, err)
		return
	}
	defer reader.Close()
	w.Header().Set("Content-Type", firstNonEmpty(info.ContentType, doc.ContentType))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, strings.ReplaceAll(doc.FileName, `"`, "")))
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
	_, _ = io.Copy(w, reader)
}

func (s *Server) deleteDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := s.service.DeleteDocument(r.Context(), currentUser(r), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getVehicleProfile(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	profile, err := s.service.GetVehicleProfile(r.Context(), currentUser(r), assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (s *Server) upsertVehicleProfile(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	var req vehicleProfileRequest
	if !readJSON(w, r, &req) {
		return
	}
	profile, err := req.toDomain(assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	out, err := s.service.UpsertVehicleProfile(r.Context(), currentUser(r), profile)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) listInsurance(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	out, err := s.service.ListInsuranceRecords(r.Context(), currentUser(r), assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createInsurance(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	var req insuranceRequest
	if !readJSON(w, r, &req) {
		return
	}
	rec, err := req.toDomain(assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	out, err := s.service.CreateInsuranceRecord(r.Context(), currentUser(r), rec)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) listLicenses(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	out, err := s.service.ListLicenseRecords(r.Context(), currentUser(r), assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createLicense(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	var req licenseRequest
	if !readJSON(w, r, &req) {
		return
	}
	rec, err := req.toDomain(assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	out, err := s.service.CreateLicenseRecord(r.Context(), currentUser(r), rec)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) listEmissions(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	out, err := s.service.ListEmissionRecords(r.Context(), currentUser(r), assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createEmission(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	var req emissionRequest
	if !readJSON(w, r, &req) {
		return
	}
	rec, err := req.toDomain(assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	out, err := s.service.CreateEmissionRecord(r.Context(), currentUser(r), rec)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) listServices(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	out, err := s.service.ListServiceRecords(r.Context(), currentUser(r), assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createService(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	var req serviceRequest
	if !readJSON(w, r, &req) {
		return
	}
	rec, err := req.toDomain(assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	out, err := s.service.CreateServiceRecord(r.Context(), currentUser(r), rec)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) listFuelLogs(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	out, err := s.service.ListFuelLogs(r.Context(), currentUser(r), assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createFuelLog(w http.ResponseWriter, r *http.Request) {
	assetID, ok := pathID(w, r)
	if !ok {
		return
	}
	var req fuelRequest
	if !readJSON(w, r, &req) {
		return
	}
	rec, err := req.toDomain(assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	out, err := s.service.CreateFuelLog(r.Context(), currentUser(r), rec)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	out, err := s.service.Dashboard(r.Context(), currentUser(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) reminders(w http.ResponseWriter, r *http.Request) {
	out, err := s.service.ListReminders(r.Context(), currentUser(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) regenerateReminders(w http.ResponseWriter, r *http.Request) {
	if err := s.service.RegenerateReminders(r.Context(), currentUser(r)); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) assetReport(w http.ResponseWriter, r *http.Request) {
	s.writeSimpleAssetCSV(w, r, "assets.csv")
}

func (s *Server) warrantyReport(w http.ResponseWriter, r *http.Request) {
	assets, err := s.service.ListAssets(r.Context(), currentUser(r), domain.AssetFilters{IncludeArchived: true})
	if err != nil {
		writeError(w, err)
		return
	}
	startCSV(w, "warranties.csv")
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"Code", "Name", "Warranty Expiry", "Warranty State", "Notes"})
	for _, a := range assets {
		if a.WarrantyExpiryDate == nil {
			continue
		}
		_ = cw.Write([]string{a.Code, a.Name, formatDate(a.WarrantyExpiryDate), a.WarrantyState, a.WarrantyNotes})
	}
	cw.Flush()
}

func (s *Server) vehicleRenewalReport(w http.ResponseWriter, r *http.Request) {
	assets, err := s.service.ListAssets(r.Context(), currentUser(r), domain.AssetFilters{IncludeArchived: true})
	if err != nil {
		writeError(w, err)
		return
	}
	startCSV(w, "vehicle-renewals.csv")
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"Code", "Name", "Renewal Type", "Reference", "Cost", "Expiry Date", "Notes"})
	for _, a := range assets {
		if a.Type != domain.AssetTypeVehicle {
			continue
		}
		insurance, _ := s.service.ListInsuranceRecords(r.Context(), currentUser(r), a.ID)
		for _, rec := range insurance {
			_ = cw.Write([]string{a.Code, a.Name, "insurance", rec.PolicyNumber, formatFloat(rec.Cost), rec.ExpiryDate.Format(dateLayout), rec.Notes})
		}
		licenses, _ := s.service.ListLicenseRecords(r.Context(), currentUser(r), a.ID)
		for _, rec := range licenses {
			_ = cw.Write([]string{a.Code, a.Name, "license", rec.ReferenceNumber, formatFloat(rec.Cost), rec.ExpiryDate.Format(dateLayout), rec.Notes})
		}
		emissions, _ := s.service.ListEmissionRecords(r.Context(), currentUser(r), a.ID)
		for _, rec := range emissions {
			_ = cw.Write([]string{a.Code, a.Name, "emission", rec.ReferenceNumber, formatFloat(rec.Cost), rec.ExpiryDate.Format(dateLayout), rec.Notes})
		}
	}
	cw.Flush()
}

func (s *Server) serviceReport(w http.ResponseWriter, r *http.Request) {
	startCSV(w, "service-history.csv")
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"Asset ID", "Date", "Type", "Vendor", "Cost", "Description"})
	assets, err := s.service.ListAssets(r.Context(), currentUser(r), domain.AssetFilters{IncludeArchived: true})
	if err != nil {
		writeError(w, err)
		return
	}
	for _, a := range assets {
		records, err := s.service.ListServiceRecords(r.Context(), currentUser(r), a.ID)
		if err != nil {
			continue
		}
		for _, rec := range records {
			_ = cw.Write([]string{a.Code, rec.ServiceDate.Format(dateLayout), rec.ServiceType, rec.Vendor, formatFloat(rec.Cost), rec.Description})
		}
	}
	cw.Flush()
}

func (s *Server) fuelReport(w http.ResponseWriter, r *http.Request) {
	startCSV(w, "fuel-logs.csv")
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"Asset ID", "Date", "Fuel Type", "Quantity", "Cost", "Odometer"})
	assets, err := s.service.ListAssets(r.Context(), currentUser(r), domain.AssetFilters{IncludeArchived: true})
	if err != nil {
		writeError(w, err)
		return
	}
	for _, a := range assets {
		if a.Type != domain.AssetTypeVehicle {
			continue
		}
		records, err := s.service.ListFuelLogs(r.Context(), currentUser(r), a.ID)
		if err != nil {
			continue
		}
		for _, rec := range records {
			_ = cw.Write([]string{a.Code, rec.FuelDate.Format(dateLayout), rec.FuelType, strconv.FormatFloat(rec.Quantity, 'f', 3, 64), strconv.FormatFloat(rec.Cost, 'f', 2, 64), formatInt(rec.Odometer)})
		}
	}
	cw.Flush()
}

func (s *Server) writeSimpleAssetCSV(w http.ResponseWriter, r *http.Request, filename string) {
	assets, err := s.service.ListAssets(r.Context(), currentUser(r), domain.AssetFilters{IncludeArchived: true})
	if err != nil {
		writeError(w, err)
		return
	}
	startCSV(w, filename)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"Code", "Name", "Type", "Category", "Status", "Location", "Assigned To", "Warranty Expiry", "Documents"})
	for _, a := range assets {
		_ = cw.Write([]string{a.Code, a.Name, a.Type, a.CategoryName, a.Status, a.Location, a.AssignedTo, formatDate(a.WarrantyExpiryDate), strconv.Itoa(a.DocumentCount)})
	}
	cw.Flush()
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := service.BearerToken(r)
		if token == "" {
			writeError(w, domain.ErrUnauthorized)
			return
		}
		user, err := s.service.UserFromToken(r.Context(), token)
		if err != nil {
			writeError(w, err)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
	}
}

func (s *Server) cors(next http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, origin := range s.cfg.CORSAllowedOrigins {
		allowed[origin] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Error("panic recovered", "panic", rec)
				writeError(w, errors.New("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func currentUser(r *http.Request) domain.User {
	user, _ := r.Context().Value(userKey).(domain.User)
	return user
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON body", domain.ErrInvalid))
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		code = "unauthorized"
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		code = "forbidden"
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, domain.ErrInvalid):
		status = http.StatusBadRequest
		code = "invalid_request"
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
		code = "conflict"
	}
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": err.Error(),
		},
	})
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, fmt.Errorf("%w: invalid id", domain.ErrInvalid))
		return 0, false
	}
	return id, true
}

func startCSV(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return "application/octet-stream"
}

func filepathExt(name string) string {
	idx := strings.LastIndex(name, ".")
	if idx < 0 {
		return ""
	}
	return name[idx:]
}
