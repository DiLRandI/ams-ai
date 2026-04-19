# Product Requirements Document (PRD)

## Project Title

Asset Management System (AMS) for Personal and Small Office Use

## Version

v0.1 (Initial PRD)

## Status

Draft

## Owner

Product / Founder

---

## 1. Overview

The Asset Management System (AMS) is a web-based application designed to help individuals and small offices track physical assets, related documents, warranty periods, service history, and recurring vehicle costs.

The system is intended to solve a common problem: people and small teams own valuable devices, appliances, furniture, and vehicles, but often lose track of purchase bills, warranty expiry dates, service records, insurance renewals, and usage history.

AMS will provide a simple, searchable system where users can:

- register and manage assets
- attach bills, invoices, warranty documents, insurance papers, and receipts
- track expiry dates and renewal dates
- record maintenance and service events
- manage vehicle-specific costs such as fuel, insurance, licensing, and service
- receive reminders before important dates are missed

The MVP will **not include AI integrations**. The system will be designed so AI features can be added later.

---

## 2. Problem Statement

People and small offices often manage assets informally using memory, paper files, spreadsheets, or messaging apps. This causes several issues:

- purchase bills and warranty cards are misplaced
- users forget warranty expiry and renewal dates
- vehicle insurance, license, and service dates are not tracked consistently
- assets are hard to search when needed
- service and operating costs are not visible over time
- there is no single place to store asset details and supporting documents

As a result, users lose money, miss renewals, fail to claim warranties, and waste time looking for records.

---

## 3. Product Vision

Create a simple and reliable asset management system that helps users keep every asset record, document, expiry date, and operating cost in one place.

The product should be especially valuable for users who want to:

- quickly search a device and see its warranty expiry
- attach a photo of a bill instead of storing paper copies
- track vehicle costs and renewal dates
- receive reminders before insurance, license, warranty, or service deadlines

---

## 4. Target Users

### Primary Users

1. **Personal users / households**
   - want to track home electronics, appliances, furniture, vehicles, and important documents
   - often lose bills and want a digital record
   - need reminders for warranties, insurance, licenses, and service

2. **Small office admins / business owners**
   - manage office laptops, printers, routers, furniture, and vehicles
   - need assignment, location, and status tracking
   - need maintenance and expiry visibility

### Secondary Users

- office staff who need to view assigned assets
- family members who may use household assets or vehicles

---

## 5. Goals

### Business Goals

- provide a practical MVP that solves real asset tracking problems without unnecessary complexity
- support both personal and small-office use cases from the start
- create a foundation for future mobile and AI-enabled versions

### Product Goals

- allow users to register and search assets quickly
- allow users to attach supporting documents to each asset
- make warranty and renewal dates highly visible
- track vehicle-specific operating costs and document expiry dates
- provide useful reminders and status visibility

### User Goals

- “I want to search an asset and instantly know when the warranty expires.”
- “I want to attach my bill photo so I do not lose it.”
- “I want to know when my vehicle insurance and license expire.”
- “I want to see how much I spend on service, fuel, and renewals.”

---

## 6. Non-Goals for MVP

The following are intentionally out of scope for the MVP:

- AI-based document extraction
- AI assistant / chat search
- OCR-based field autofill
- predictive maintenance
- advanced accounting integrations
- multi-branch enterprise workflows
- procurement workflows / approvals
- native mobile apps
- advanced depreciation modeling

These may be considered in later phases.

---

## 7. MVP Scope

### In Scope

#### 7.1 Asset Management

- create, edit, archive, and view assets
- support asset categories
- assign unique asset identifiers
- store basic asset details
- search and filter assets

#### 7.2 Asset Categories

- general devices and equipment
- appliances
- furniture
- networking equipment
- tools
- vehicles
- custom/general categories as needed

#### 7.3 Document Attachment

- upload and attach files to assets
- support bill/invoice images
- support PDF documents
- support multiple documents per asset
- categorize documents (bill, warranty, insurance, service receipt, manual, other)

#### 7.4 Warranty & Expiry Tracking

- record warranty start and expiry dates
- display warranty status on asset detail and search/detail views
- show whether warranty is active, expiring soon, or expired

#### 7.5 Vehicle Management

- create vehicle assets as a dedicated asset type
- store vehicle-specific details
- record insurance cost and expiry date
- record license cost and expiry date
- record emission/inspection cost and expiry date if needed
- record service history and service due date or mileage
- record fuel logs and fuel cost

#### 7.6 Maintenance & Service Tracking

- add service/repair records to assets
- record service date, cost, vendor, and notes
- view service history per asset

#### 7.7 Reminders / Notifications

- upcoming warranty expiry reminders
- vehicle insurance expiry reminders
- vehicle license expiry reminders
- service due reminders

#### 7.8 Basic Dashboard

- total assets
- assets nearing warranty expiry
- vehicles nearing insurance/license expiry
- recently added assets
- assets needing service soon

---

## 8. Core User Flows

### 8.1 Add a General Asset

1. User opens Add Asset screen
2. User selects category
3. User enters asset details
4. User adds warranty details if available
5. User uploads bill photo / invoice / documents
6. User saves asset

### 8.2 Search for an Asset

1. User searches by asset name, model, serial number, or registration number
2. System returns matching asset(s)
3. User opens asset detail
4. User views warranty expiry, documents, and service history

### 8.3 Add a Vehicle

1. User creates an asset with type = vehicle
2. User enters registration and vehicle details
3. User adds insurance cost and expiry
4. User adds license cost and expiry
5. User optionally adds service and fuel history
6. User saves vehicle asset

### 8.4 Attach Bill or Document

1. User opens an asset detail page
2. User uploads one or more files
3. User selects document type
4. System stores and links document to the asset
5. User can later open/download/view the file

### 8.5 Record Service or Renewal

1. User opens asset or vehicle detail page
2. User adds service or renewal record
3. User enters date, cost, notes, and vendor if applicable
4. System updates history and relevant due/expiry status

---

## 9. Functional Requirements

### 9.1 Asset Records

The system shall:

- allow users to create and edit asset records
- store asset name, category, model, brand, serial number, purchase date, purchase price, location, status, and notes
- support optional warranty dates and document attachments
- support asset statuses such as active, in repair, stored, retired, disposed

### 9.2 Search & Filtering

The system shall:

- allow users to search by asset name, model, serial number, category, and vehicle registration number
- allow filtering by category, status, warranty state, and document availability
- show key summary information in search results

### 9.3 Warranty Tracking

The system shall:

- store warranty expiry date for applicable assets
- display expiry date on asset summary and detail views
- visually indicate whether the warranty is active, near expiry, or expired

### 9.4 Document Management

The system shall:

- support upload of images and PDFs
- allow multiple documents to be attached to one asset
- store document type and upload date
- allow viewing/downloading attached documents

### 9.5 Vehicle Management

The system shall:

- support vehicle as a distinct asset type
- store vehicle registration number, brand/model, purchase date, purchase price, and notes
- store insurance cost and expiry date
- store license cost and expiry date
- store emission/inspection cost and expiry date if configured
- allow service entries and fuel log entries to be attached to the vehicle

### 9.6 Service Records

The system shall:

- allow users to create service/repair entries
- store date, cost, vendor, notes, and related asset
- show service history in chronological order

### 9.7 Fuel Logs

The system shall:

- allow users to record fuel date, quantity, cost, and odometer reading
- show fuel history for each vehicle

### 9.8 Reminders

The system shall:

- identify upcoming expiry dates based on configurable threshold windows
- show reminder items in the dashboard
- support future email notification capability

---

## 10. Non-Functional Requirements

### 10.1 Usability

- the UI should be simple enough for non-technical users
- common tasks should require minimal steps
- the product should work well on desktop and mobile browsers

### 10.2 Performance

- asset list/search responses should feel fast for small to medium data volumes
- uploads should complete reliably for normal document sizes

### 10.3 Security

- only authenticated users can access asset records
- permissions should support at least admin and regular user roles
- uploaded documents should not be publicly exposed

### 10.4 Reliability

- asset and document data should be stored persistently
- reminder logic should run consistently
- backups should be possible for database and documents

### 10.5 Maintainability

- product should be structured so mobile apps can use the same backend later
- API contracts should remain stable and documented

---

## 11. Data Entities (High-Level)

The MVP is expected to include at least these entities:

- User
- Role
- Asset
- AssetCategory
- AssetDocument
- Warranty
- VehicleProfile
- VehicleInsuranceRecord
- VehicleLicenseRecord
- VehicleEmissionRecord (optional in MVP depending on region)
- ServiceRecord
- FuelLog
- Reminder

---

## 12. Key Screens

### Web MVP Screens

- login
- dashboard
- asset list
- asset detail
- add/edit asset
- document upload/view section
- vehicle detail
- add service record
- add fuel log
- reminders / upcoming expiries view

---

## 13. Sample Asset Fields

### General Asset

- asset name
- category
- brand
- model
- serial number
- purchase date
- purchase cost
- warranty expiry date
- location
- assigned user (optional)
- status
- notes

### Vehicle Asset

- asset name
- registration number
- brand/model
- purchase date
- purchase cost
- odometer
- insurance cost
- insurance expiry date
- license cost
- license expiry date
- emission/inspection cost
- emission/inspection expiry date
- next service date or mileage
- notes

---

## 14. User Stories

### Personal User

- As a personal user, I want to register my electronics and attach their bills so I do not lose proof of purchase.
- As a personal user, I want to search a device and immediately see the warranty expiry date.
- As a personal user, I want to track insurance and license expiry dates for my vehicle.
- As a personal user, I want to view my vehicle’s service and fuel history.

### Small Office Admin

- As an office admin, I want to store all office asset details in one system.
- As an office admin, I want to attach invoices and service receipts to assets.
- As an office admin, I want to track warranties and upcoming expiry dates.
- As an office admin, I want to manage company vehicles and their renewal dates.

---

## 15. Acceptance Criteria for MVP

The MVP is successful when:

- users can create and manage asset records
- users can attach one or more documents to an asset
- users can search for an asset and view warranty expiry information
- users can create vehicle assets with insurance and license cost/expiry data
- users can record service and fuel entries for vehicles
- users can view a dashboard of upcoming expiries and due items

---

## 16. Success Metrics

### Product Success Metrics

- number of assets created
- percentage of assets with attached documents
- percentage of applicable assets with warranty dates recorded
- number of vehicle records with insurance/license dates recorded
- number of reminder-triggering items tracked in system

### User Value Metrics

- users can locate asset records quickly
- users reduce missed warranty/renewal deadlines
- users reduce lost bill/invoice incidents by storing documents digitally

---

## 17. Risks and Assumptions

### Assumptions

- users are willing to enter asset information manually in MVP
- users value document storage and reminder visibility enough to adopt the system
- vehicle management is important for both personal and small-office use

### Risks

- users may not consistently upload documents unless the flow is very easy
- reminder logic may lose value if dates are not entered accurately
- search quality may feel weak if data is incomplete or inconsistent

---

## 18. Future Enhancements

Potential future enhancements after MVP:

- AI-based bill parsing and document extraction
- OCR for serial numbers and invoices
- natural language search
- React Native mobile app
- barcode/QR scanning from mobile devices
- email and push notifications
- advanced reporting and cost analytics
- depreciation and resale value tracking
- multi-office / branch support

---

## 19. Suggested Release Phases

### Phase 1: MVP

- core asset management
- warranty tracking
- document attachment
- vehicle insurance/license/service/fuel tracking
- reminders dashboard

### Phase 2

- stronger reporting
- exports
- assignment workflows
- better notification channels
- mobile optimization improvements

### Phase 3

- mobile app
- AI-assisted document extraction
- smart recommendations

---

## 20. Summary

AMS is a practical asset management platform focused on keeping records, documents, expiry dates, and recurring costs organized for personal and small-office use.

The MVP should focus on solving the most immediate and valuable problems:

- keeping asset records organized
- preserving bills and supporting documents
- making warranty and renewal dates visible
- managing vehicle costs and document expiries
- giving users a simple searchable system instead of paper files and scattered spreadsheets
