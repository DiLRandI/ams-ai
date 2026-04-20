export type Role = "admin" | "user";
export type AssetType = "general" | "vehicle";
export type AssetStatus =
  | "active"
  | "in_repair"
  | "stored"
  | "retired"
  | "disposed";
export type WarrantyState = "active" | "expiring_soon" | "expired" | "not_set";

export interface User {
  id: number;
  email: string;
  fullName: string;
  role: Role;
}

export interface Category {
  id: number;
  name: string;
  description: string;
  isSystem: boolean;
}

export interface Asset {
  id: number;
  code: string;
  type: AssetType;
  categoryId: number;
  categoryName?: string;
  name: string;
  brand: string;
  model: string;
  serialNumber: string;
  purchaseDate?: string;
  purchasePrice?: number;
  status: AssetStatus;
  condition: string;
  location: string;
  assignedTo: string;
  assignedUserId?: number;
  assignedUserName?: string;
  notes: string;
  warrantyStartDate?: string;
  warrantyExpiryDate?: string;
  warrantyNotes: string;
  archivedAt?: string;
  documentCount?: number;
  warrantyState?: WarrantyState;
  createdAt: string;
  updatedAt: string;
}

export interface VehicleProfile {
  assetId: number;
  registrationNumber: string;
  vehicleType: string;
  chassisNumber: string;
  engineNumber: string;
  odometer?: number;
  assignedDriver: string;
  nextServiceDate?: string;
  nextServiceMileage?: number;
  notes: string;
}

export interface VehicleInsuranceRecord {
  id: number;
  assetId: number;
  provider: string;
  policyNumber: string;
  cost?: number;
  startDate?: string;
  expiryDate: string;
  documentId?: number;
  notes: string;
}

export interface VehicleLicenseRecord {
  id: number;
  assetId: number;
  renewalType: string;
  referenceNumber: string;
  cost?: number;
  issueDate?: string;
  expiryDate: string;
  documentId?: number;
  notes: string;
}

export interface VehicleEmissionRecord {
  id: number;
  assetId: number;
  inspectionType: string;
  referenceNumber: string;
  cost?: number;
  issueDate?: string;
  expiryDate: string;
  documentId?: number;
  notes: string;
}

export interface ServiceRecord {
  id: number;
  assetId: number;
  serviceType: "service" | "repair";
  serviceDate: string;
  cost?: number;
  vendor: string;
  description: string;
  notes: string;
  mileage?: number;
  nextServiceDate?: string;
  nextServiceMileage?: number;
}

export interface FuelLog {
  id: number;
  assetId: number;
  fuelDate: string;
  fuelType: string;
  quantity: number;
  cost: number;
  odometer?: number;
  notes: string;
}

export interface AssetDocument {
  id: number;
  assetId: number;
  title: string;
  type: string;
  notes: string;
  fileName: string;
  contentType: string;
  sizeBytes: number;
  uploadedBy: number;
  createdAt: string;
}

export interface Reminder {
  id: number;
  assetId: number;
  assetCode?: string;
  assetName?: string;
  sourceType: string;
  sourceId: number;
  title: string;
  dueDate: string;
  state: "upcoming" | "due" | "overdue";
}

export interface Dashboard {
  totalAssets: number;
  assetsByCategory: Array<{
    categoryId: number;
    categoryName: string;
    count: number;
  }>;
  expiringWarranties: Asset[];
  expiringVehicleInsurance: Reminder[];
  expiringVehicleLicenses: Reminder[];
  recentlyAddedAssets: Asset[];
  serviceDueSoon: Reminder[];
  upcomingReminders: Reminder[];
}

export interface LoginResponse {
  token: string;
  expiresAt: string;
  user: User;
}
