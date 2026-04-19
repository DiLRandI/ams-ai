import type { AssetStatus, WarrantyState } from '../api/types';

const statusLabels: Record<AssetStatus, string> = {
  active: 'Active',
  in_repair: 'In repair',
  stored: 'Stored',
  retired: 'Retired',
  disposed: 'Disposed'
};

const warrantyLabels: Record<WarrantyState, string> = {
  active: 'Active',
  expiring_soon: 'Expiring soon',
  expired: 'Expired',
  not_set: 'Not set'
};

export function StatusBadge({ status }: { status: AssetStatus }) {
  return <span className={`badge status-${status}`}>{statusLabels[status] ?? status}</span>;
}

export function WarrantyBadge({ state }: { state?: WarrantyState }) {
  const safeState = state ?? 'not_set';
  return <span className={`badge warranty-${safeState}`}>{warrantyLabels[safeState]}</span>;
}

export function ReminderBadge({ state }: { state: string }) {
  return <span className={`badge reminder-${state}`}>{state}</span>;
}
