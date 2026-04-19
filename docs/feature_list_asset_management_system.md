# Feature List Document

## Project Title

Asset Management System (AMS) for Personal and Small Office Use

## Version

v0.1

## Status

Draft

---

## 1. Purpose

This document lists and prioritizes the features for the Asset Management System (AMS). It is intended to complement the PRD and provide a practical build roadmap.

The feature list is grouped into:

- MVP features
- Post-MVP / Phase 2 features
- Future / Phase 3 features
- Explicit out-of-scope items for MVP

---

## 2. Product Scope Summary

AMS is designed to help personal users and small offices:

- track assets and their ownership/use
- attach bills, invoices, warranty documents, and other records
- monitor expiry dates and renewal dates
- manage vehicle-specific costs and renewals
- record maintenance and service history
- search assets quickly and view important status information

The MVP will not include AI integrations.

---

## 3. MVP Features

## 3.1 Authentication and User Access

### Core Features

- User login/logout
- Email/password authentication
- Basic user roles
  - Admin
  - Standard user
- User profile management

### Notes

- Admin can manage assets, documents, reminders, and settings
- Standard users can view and update only permitted records

---

## 3.2 Asset Management

### Core Features

- Create asset
- Edit asset
- View asset detail
- Archive asset
- Restore archived asset if needed
- Unique asset ID/code

### Asset Fields

- Asset name
- Category
- Brand
- Model
- Serial number
- Purchase date
- Purchase price
- Status
- Location
- Assigned user/person
- Notes

### Asset Statuses

- Active
- In repair
- Stored
- Retired
- Disposed

---

## 3.3 Asset Categories

### Core Features

- Predefined categories
- Category-based filtering
- Ability to add/edit categories in admin settings

### Initial Categories

- IT devices
- Networking equipment
- Appliances
- Furniture
- Tools
- Vehicles
- Other / General

---

## 3.4 Search and Filtering

### Core Features

- Global asset search
- Filter assets by category
- Filter by status
- Filter by location
- Filter by assigned user
- Filter by warranty state
- Filter by document availability

### Searchable Fields

- Asset name
- Brand
- Model
- Serial number
- Notes
- Vehicle registration number

### Search Result Summary

Each asset search result should show:

- asset name
- category
- status
- location or assigned user
- warranty expiry date if available
- whether documents are attached

---

## 3.5 Document Management

### Core Features

- Upload documents to an asset
- View/download attached documents
- Delete or replace documents
- Support multiple documents per asset
- Document type selection

### Supported Document Types

- Bill / invoice
- Warranty document
- Insurance document
- License / registration document
- Service receipt
- Manual
- Other

### Supported File Types

- JPG / JPEG
- PNG
- PDF

### Document Metadata

- Document title
- Document type
- Upload date
- Optional notes

---

## 3.6 Warranty and Expiry Tracking

### Core Features

- Warranty start date
- Warranty expiry date
- Warranty notes
- Warranty status visibility on list/detail/search views

### Warranty States

- Active
- Expiring soon
- Expired
- Not set

### Reminder Support

- Notify before warranty expiry
- Default reminder windows configurable later

---

## 3.7 Vehicle Asset Management

### Core Features

- Dedicated vehicle asset type
- Vehicle-specific detail section
- Track renewal and operating records

### Vehicle Fields

- Registration number
- Vehicle type
- Brand/model
- Chassis number
- Engine number
- Purchase date
- Purchase price
- Current mileage / odometer
- Assigned driver/user
- Notes

---

## 3.8 Vehicle Insurance Tracking

### Core Features

- Insurance provider
- Policy/reference number
- Insurance cost
- Insurance start date
- Insurance expiry date
- Attach insurance document

### Reminder Support

- Notify before insurance expiry

---

## 3.9 Vehicle License / Renewal Tracking

### Core Features

- License/renewal type
- License cost
- License issue/renewal date
- License expiry date
- Attach related document

### Reminder Support

- Notify before license expiry

---

## 3.10 Vehicle Emission / Inspection Tracking

### Core Features

- Emission/inspection record type
- Cost
- Issue date
- Expiry date
- Attach certificate/document

### Reminder Support

- Notify before expiry

---

## 3.11 Service and Maintenance Tracking

### Core Features

- Add service record to any asset
- Add repair record to any asset
- View service/repair history

### Service Record Fields

- Date
- Cost
- Vendor/service provider
- Description
- Notes
- Current mileage (for vehicle service if relevant)
- Next service date and/or next service mileage

### Reminder Support

- Notify before next service due date
- Vehicle service due by mileage can be shown as a manual status in MVP

---

## 3.12 Fuel Log Tracking

### Core Features

- Add fuel entry for vehicles
- View fuel history per vehicle

### Fuel Entry Fields

- Date
- Fuel type
- Quantity
- Cost
- Odometer reading
- Notes

### MVP Output

- Fuel cost history per vehicle
- Total fuel cost on vehicle detail screen

---

## 3.13 Dashboard

### Core Features

- Total asset count
- Assets by category
- Assets with expiring warranties
- Vehicles with expiring insurance
- Vehicles with expiring licenses
- Recently added assets
- Assets/vehicles needing service soon

---

## 3.14 Notifications and Reminders

### Core Features

- In-app reminder list
- Email reminders (optional if included in MVP build plan)
- Reminder generation for expiry-based records

### Reminder Types

- Warranty expiry
- Insurance expiry
- License expiry
- Emission/inspection expiry
- Service due date

---

## 3.15 Basic Reporting

### Core Features

- Asset list export
- Warranty expiry report
- Vehicle renewal report
- Service history report
- Fuel log export

### Export Formats

- CSV
- PDF optional later if needed

---

## 4. Phase 2 Features

## 4.1 Improved Assignment and Ownership

- Asset handover history
- Check-in / check-out flow
- Assignment change log
- Family member / employee ownership view

## 4.2 Better Dashboard and Reporting

- Monthly cost charts
- Vehicle annual cost summary
- Cost by category
- Cost by vendor/service provider
- Spending trends

## 4.3 Better Document Features

- Document preview inside the app
- Bulk upload
- Drag-and-drop upload
- Document tagging
- Search documents directly

## 4.4 Richer Notifications

- Multiple reminder intervals
- Repeating reminders until updated
- Reminder preferences per user

## 4.5 Data Quality Features

- Required-field validation improvements
- Duplicate asset detection rules
- Better import validation

## 4.6 Bulk Operations

- Bulk asset import from CSV
- Bulk status updates
- Bulk category reassignment

## 4.7 More Vehicle Features

- Tire replacement history
- Battery replacement history
- Parking/toll/other expense logs
- Cost-per-period summaries

---

## 5. Phase 3 / Future Features

## 5.1 Mobile App

- React Native mobile application
- Mobile document upload
- Mobile asset lookup
- Quick vehicle log entry
- Push notifications

## 5.2 QR / Barcode Support

- Generate QR code for each asset
- Scan QR code to open asset detail
- Printable labels

## 5.3 Advanced Reporting

- Visual analytics dashboards
- More export formats
- Scheduled report generation

## 5.4 AI Features (Future)

- Bill/document field extraction
- OCR-assisted data entry
- Smart search assistant
- Duplicate detection assistance
- Reminder suggestions
- Basic maintenance recommendations

## 5.5 Integrations

- Email service integrations
- Calendar reminders
- Accounting/export integrations
- Cloud drive sync for documents

---

## 6. Out of Scope for MVP

The following are not planned for MVP:

- AI or OCR features
- Native mobile app
- Barcode/QR scanning workflows
- Procurement workflows
- Approval chains
- Multi-branch enterprise support
- Accounting integrations
- Advanced depreciation engine
- Predictive maintenance
- Real-time collaboration features

---

## 7. MVP Prioritization Summary

## Must Have

- Authentication
- Asset CRUD
- Categories
- Search and filtering
- Document attachments
- Warranty tracking
- Vehicle records
- Insurance/license/emission expiry tracking
- Service records
- Fuel logs
- Dashboard
- Reminders
- Basic reporting/export

## Should Have

- Assignment fields
- Archive/restore flow
- Email reminders
- CSV export

## Could Have

- CSV import
- Better document preview
- Assignment history
- More advanced vehicle expense breakdowns

---

## 8. Notes for Engineering

### Suggested Build Order

1. Authentication and user roles
2. Asset and category management
3. Document attachment module
4. Warranty and expiry module
5. Vehicle management
6. Service and fuel logs
7. Dashboard
8. Reminder engine
9. Reporting/export

### Suggested Principle

Build the MVP around the main user promise:

**Track every asset, keep its documents safe, and never miss important expiry or service dates.**

---

## 9. Open Questions

- Should email reminders be part of MVP or first post-MVP enhancement?
- Should CSV import be MVP or Phase 2?
- Should asset assignment history be included in MVP?
- Should reminder intervals be fixed initially or user-configurable?
- Should PDF export be MVP or Phase 2?
