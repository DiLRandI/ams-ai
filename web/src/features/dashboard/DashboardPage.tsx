import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Bell, Car, Clock, Package, ShieldCheck, Wrench } from "lucide-react";
import { api } from "../../api/client";
import { PageHeader } from "../../components/PageHeader";
import { EmptyState } from "../../components/EmptyState";
import { Loading } from "../../components/Loading";
import {
  WarrantyBadge,
  ReminderBadge,
  StatusBadge,
} from "../../components/StateBadge";
import { dateOnly } from "../assets/format";

export function DashboardPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["dashboard"],
    queryFn: api.dashboard,
  });

  if (isLoading) return <Loading />;
  if (error)
    return (
      <div className="alert">
        {error instanceof Error ? error.message : "Dashboard failed"}
      </div>
    );
  if (!data) return null;

  return (
    <>
      <PageHeader title="Dashboard" eyebrow="MVP overview" />
      <section className="metricGrid">
        <div className="metric">
          <Package size={22} />
          <strong>{data.totalAssets}</strong>
          <span>Total assets</span>
        </div>
        <div className="metric">
          <ShieldCheck size={22} />
          <strong>{data.expiringWarranties.length}</strong>
          <span>Warranties expiring</span>
        </div>
        <div className="metric">
          <Car size={22} />
          <strong>
            {data.expiringVehicleInsurance.length +
              data.expiringVehicleLicenses.length}
          </strong>
          <span>Vehicle renewals</span>
        </div>
        <div className="metric">
          <Wrench size={22} />
          <strong>{data.serviceDueSoon.length}</strong>
          <span>Service due</span>
        </div>
      </section>

      <section className="gridTwo">
        <div className="panel">
          <h2>Upcoming reminders</h2>
          {data.upcomingReminders.length === 0 ? (
            <EmptyState title="No upcoming reminders" />
          ) : (
            <ul className="compactList">
              {data.upcomingReminders.slice(0, 8).map((item) => (
                <li key={item.id}>
                  <Bell size={16} />
                  <div>
                    <Link to={`/assets/${item.assetId}`}>{item.title}</Link>
                    <span>{dateOnly(item.dueDate)}</span>
                  </div>
                  <ReminderBadge state={item.state} />
                </li>
              ))}
            </ul>
          )}
        </div>
        <div className="panel">
          <h2>Assets by category</h2>
          <div className="categoryBars">
            {data.assetsByCategory.map((item) => (
              <div key={item.categoryId}>
                <span>{item.categoryName}</span>
                <strong>{item.count}</strong>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="gridTwo">
        <div className="panel">
          <h2>Recently added</h2>
          {data.recentlyAddedAssets.length === 0 ? (
            <EmptyState
              title="No assets yet"
              body="Create the first asset from the assets screen."
            />
          ) : (
            <ul className="compactList">
              {data.recentlyAddedAssets.slice(0, 6).map((asset) => (
                <li key={asset.id}>
                  <Package size={16} />
                  <div>
                    <Link to={`/assets/${asset.id}`}>{asset.name}</Link>
                    <span>{asset.code}</span>
                  </div>
                  <StatusBadge status={asset.status} />
                </li>
              ))}
            </ul>
          )}
        </div>
        <div className="panel">
          <h2>Warranty attention</h2>
          {data.expiringWarranties.length === 0 ? (
            <EmptyState title="No warranty issues" />
          ) : (
            <ul className="compactList">
              {data.expiringWarranties.map((asset) => (
                <li key={asset.id}>
                  <Clock size={16} />
                  <div>
                    <Link to={`/assets/${asset.id}`}>{asset.name}</Link>
                    <span>{dateOnly(asset.warrantyExpiryDate)}</span>
                  </div>
                  <WarrantyBadge state={asset.warrantyState} />
                </li>
              ))}
            </ul>
          )}
        </div>
      </section>
    </>
  );
}
