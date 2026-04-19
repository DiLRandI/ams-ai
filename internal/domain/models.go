package domain

import (
	"errors"
	"strings"
	"time"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"

	AssetTypeGeneral = "general"
	AssetTypeVehicle = "vehicle"

	StatusActive   = "active"
	StatusInRepair = "in_repair"
	StatusStored   = "stored"
	StatusRetired  = "retired"
	StatusDisposed = "disposed"

	WarrantyActive       = "active"
	WarrantyExpiringSoon = "expiring_soon"
	WarrantyExpired      = "expired"
	WarrantyNotSet       = "not_set"

	ReminderUpcoming = "upcoming"
	ReminderDue      = "due"
	ReminderOverdue  = "overdue"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalid      = errors.New("invalid input")
	ErrConflict     = errors.New("conflict")
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"fullName"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Category struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"isSystem"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Asset struct {
	ID                 int64      `json:"id"`
	Code               string     `json:"code"`
	Type               string     `json:"type"`
	CategoryID         int64      `json:"categoryId"`
	CategoryName       string     `json:"categoryName,omitempty"`
	Name               string     `json:"name"`
	Brand              string     `json:"brand"`
	Model              string     `json:"model"`
	SerialNumber       string     `json:"serialNumber"`
	PurchaseDate       *time.Time `json:"purchaseDate,omitempty"`
	PurchasePrice      *float64   `json:"purchasePrice,omitempty"`
	Status             string     `json:"status"`
	Condition          string     `json:"condition"`
	Location           string     `json:"location"`
	AssignedTo         string     `json:"assignedTo"`
	AssignedUserID     *int64     `json:"assignedUserId,omitempty"`
	AssignedUserName   string     `json:"assignedUserName,omitempty"`
	Notes              string     `json:"notes"`
	WarrantyStartDate  *time.Time `json:"warrantyStartDate,omitempty"`
	WarrantyExpiryDate *time.Time `json:"warrantyExpiryDate,omitempty"`
	WarrantyNotes      string     `json:"warrantyNotes"`
	ArchivedAt         *time.Time `json:"archivedAt,omitempty"`
	CreatedBy          int64      `json:"createdBy"`
	UpdatedBy          *int64     `json:"updatedBy,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	DocumentCount      int        `json:"documentCount,omitempty"`
	WarrantyState      string     `json:"warrantyState,omitempty"`
}

type AssetFilters struct {
	Query             string
	CategoryID        int64
	Status            string
	Location          string
	AssignedUserID    int64
	WarrantyState     string
	HasDocuments      *bool
	IncludeArchived   bool
	CurrentUserID     int64
	CurrentUserRole   string
	ReminderWindowDay int
}

type VehicleProfile struct {
	AssetID            int64      `json:"assetId"`
	RegistrationNumber string     `json:"registrationNumber"`
	VehicleType        string     `json:"vehicleType"`
	ChassisNumber      string     `json:"chassisNumber"`
	EngineNumber       string     `json:"engineNumber"`
	Odometer           *int       `json:"odometer,omitempty"`
	AssignedDriver     string     `json:"assignedDriver"`
	NextServiceDate    *time.Time `json:"nextServiceDate,omitempty"`
	NextServiceMileage *int       `json:"nextServiceMileage,omitempty"`
	Notes              string     `json:"notes"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type VehicleInsuranceRecord struct {
	ID           int64      `json:"id"`
	AssetID      int64      `json:"assetId"`
	Provider     string     `json:"provider"`
	PolicyNumber string     `json:"policyNumber"`
	Cost         *float64   `json:"cost,omitempty"`
	StartDate    *time.Time `json:"startDate,omitempty"`
	ExpiryDate   time.Time  `json:"expiryDate"`
	DocumentID   *int64     `json:"documentId,omitempty"`
	Notes        string     `json:"notes"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type VehicleLicenseRecord struct {
	ID              int64      `json:"id"`
	AssetID         int64      `json:"assetId"`
	RenewalType     string     `json:"renewalType"`
	ReferenceNumber string     `json:"referenceNumber"`
	Cost            *float64   `json:"cost,omitempty"`
	IssueDate       *time.Time `json:"issueDate,omitempty"`
	ExpiryDate      time.Time  `json:"expiryDate"`
	DocumentID      *int64     `json:"documentId,omitempty"`
	Notes           string     `json:"notes"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type VehicleEmissionRecord struct {
	ID              int64      `json:"id"`
	AssetID         int64      `json:"assetId"`
	InspectionType  string     `json:"inspectionType"`
	ReferenceNumber string     `json:"referenceNumber"`
	Cost            *float64   `json:"cost,omitempty"`
	IssueDate       *time.Time `json:"issueDate,omitempty"`
	ExpiryDate      time.Time  `json:"expiryDate"`
	DocumentID      *int64     `json:"documentId,omitempty"`
	Notes           string     `json:"notes"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type ServiceRecord struct {
	ID                 int64      `json:"id"`
	AssetID            int64      `json:"assetId"`
	ServiceType        string     `json:"serviceType"`
	ServiceDate        time.Time  `json:"serviceDate"`
	Cost               *float64   `json:"cost,omitempty"`
	Vendor             string     `json:"vendor"`
	Description        string     `json:"description"`
	Notes              string     `json:"notes"`
	Mileage            *int       `json:"mileage,omitempty"`
	NextServiceDate    *time.Time `json:"nextServiceDate,omitempty"`
	NextServiceMileage *int       `json:"nextServiceMileage,omitempty"`
	CreatedBy          int64      `json:"createdBy"`
	CreatedAt          time.Time  `json:"createdAt"`
}

type FuelLog struct {
	ID        int64     `json:"id"`
	AssetID   int64     `json:"assetId"`
	FuelDate  time.Time `json:"fuelDate"`
	FuelType  string    `json:"fuelType"`
	Quantity  float64   `json:"quantity"`
	Cost      float64   `json:"cost"`
	Odometer  *int      `json:"odometer,omitempty"`
	Notes     string    `json:"notes"`
	CreatedBy int64     `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}

type AssetDocument struct {
	ID          int64     `json:"id"`
	AssetID     int64     `json:"assetId"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Notes       string    `json:"notes"`
	FileName    string    `json:"fileName"`
	ContentType string    `json:"contentType"`
	SizeBytes   int64     `json:"sizeBytes"`
	ObjectKey   string    `json:"-"`
	UploadedBy  int64     `json:"uploadedBy"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Reminder struct {
	ID         int64     `json:"id"`
	AssetID    int64     `json:"assetId"`
	AssetCode  string    `json:"assetCode,omitempty"`
	AssetName  string    `json:"assetName,omitempty"`
	SourceType string    `json:"sourceType"`
	SourceID   int64     `json:"sourceId"`
	Title      string    `json:"title"`
	DueDate    time.Time `json:"dueDate"`
	State      string    `json:"state"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Dashboard struct {
	TotalAssets              int             `json:"totalAssets"`
	AssetsByCategory         []CategoryCount `json:"assetsByCategory"`
	ExpiringWarranties       []Asset         `json:"expiringWarranties"`
	ExpiringVehicleInsurance []Reminder      `json:"expiringVehicleInsurance"`
	ExpiringVehicleLicenses  []Reminder      `json:"expiringVehicleLicenses"`
	RecentlyAddedAssets      []Asset         `json:"recentlyAddedAssets"`
	ServiceDueSoon           []Reminder      `json:"serviceDueSoon"`
	UpcomingReminders        []Reminder      `json:"upcomingReminders"`
}

type CategoryCount struct {
	CategoryID   int64  `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Count        int    `json:"count"`
}

func WarrantyState(expiry *time.Time, now time.Time, windowDays int) string {
	if expiry == nil {
		return WarrantyNotSet
	}
	today := truncateDate(now)
	exp := truncateDate(*expiry)
	switch {
	case exp.Before(today):
		return WarrantyExpired
	case !exp.After(today.AddDate(0, 0, windowDays)):
		return WarrantyExpiringSoon
	default:
		return WarrantyActive
	}
}

func ReminderState(dueDate, now time.Time) string {
	today := truncateDate(now)
	due := truncateDate(dueDate)
	switch {
	case due.Before(today):
		return ReminderOverdue
	case due.Equal(today):
		return ReminderDue
	default:
		return ReminderUpcoming
	}
}

func AssetAccessAllowed(user User, asset Asset) bool {
	if user.Role == RoleAdmin {
		return true
	}
	if asset.CreatedBy == user.ID {
		return true
	}
	return asset.AssignedUserID != nil && *asset.AssignedUserID == user.ID
}

func NormalizeStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	status = strings.ReplaceAll(status, " ", "_")
	if status == "" {
		return StatusActive
	}
	return status
}

func IsValidStatus(status string) bool {
	switch status {
	case StatusActive, StatusInRepair, StatusStored, StatusRetired, StatusDisposed:
		return true
	default:
		return false
	}
}

func IsValidAssetType(assetType string) bool {
	return assetType == AssetTypeGeneral || assetType == AssetTypeVehicle
}

func IsValidRole(role string) bool {
	return role == RoleAdmin || role == RoleUser
}

func truncateDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
