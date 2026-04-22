import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Plus, Search } from "lucide-react";
import { useMemo, useState } from "react";
import { api } from "../../api/client";
import { PageHeader } from "../../components/PageHeader";
import { EmptyState } from "../../components/EmptyState";
import { Loading } from "../../components/Loading";
import { StatusBadge, WarrantyBadge } from "../../components/StateBadge";
import { dateOnly } from "./format";

export function AssetsPage() {
  const [q, setQ] = useState("");
  const [categoryId, setCategoryId] = useState("");
  const [status, setStatus] = useState("");
  const [warrantyState, setWarrantyState] = useState("");
  const [hasDocuments, setHasDocuments] = useState("");
  const { data: categories = [] } = useQuery({
    queryKey: ["categories"],
    queryFn: api.categories,
  });

  const params = useMemo(() => {
    const next = new URLSearchParams();
    if (q) next.set("q", q);
    if (categoryId) next.set("categoryId", categoryId);
    if (status) next.set("status", status);
    if (warrantyState) next.set("warrantyState", warrantyState);
    if (hasDocuments) next.set("hasDocuments", hasDocuments);
    return next;
  }, [categoryId, hasDocuments, q, status, warrantyState]);

  const { data, isLoading, error } = useQuery({
    queryKey: ["assets", params.toString()],
    queryFn: () => api.assets(params),
  });

  return (
    <>
      <PageHeader
        title="Assets"
        eyebrow="Search and manage physical assets"
        actions={
          <Link className="primaryButton" to="/assets/new">
            <Plus size={18} /> Add asset
          </Link>
        }
      />
      <section className="filterBar">
        <label className="searchBox">
          <Search size={18} />
          <input
            placeholder="Search name, model, serial, registration..."
            value={q}
            onChange={(event) => setQ(event.target.value)}
          />
        </label>
        <select
          value={categoryId}
          onChange={(event) => setCategoryId(event.target.value)}
          aria-label="Category"
        >
          <option value="">All categories</option>
          {categories.map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </select>
        <select
          value={status}
          onChange={(event) => setStatus(event.target.value)}
          aria-label="Status"
        >
          <option value="">All statuses</option>
          <option value="active">Active</option>
          <option value="in_repair">In repair</option>
          <option value="stored">Stored</option>
          <option value="retired">Retired</option>
          <option value="disposed">Disposed</option>
        </select>
        <select
          value={warrantyState}
          onChange={(event) => setWarrantyState(event.target.value)}
          aria-label="Warranty state"
        >
          <option value="">All warranties</option>
          <option value="active">Active</option>
          <option value="expiring_soon">Expiring soon</option>
          <option value="expired">Expired</option>
          <option value="not_set">Not set</option>
        </select>
        <select
          value={hasDocuments}
          onChange={(event) => setHasDocuments(event.target.value)}
          aria-label="Documents"
        >
          <option value="">Any documents</option>
          <option value="true">Has documents</option>
          <option value="false">No documents</option>
        </select>
      </section>
      <section className="panel">
        {isLoading && <Loading />}
        {error && (
          <div className="alert">
            {error instanceof Error ? error.message : "Could not load assets"}
          </div>
        )}
        {!isLoading && !data?.length && (
          <EmptyState
            title="No assets found"
            body="Try clearing filters or add a new asset."
          />
        )}
        {!!data?.length && (
          <div className="tableWrap">
            <table>
              <thead>
                <tr>
                  <th>Asset</th>
                  <th>Category</th>
                  <th>Status</th>
                  <th>Location / Assigned</th>
                  <th>Warranty</th>
                  <th>Documents</th>
                </tr>
              </thead>
              <tbody>
                {data.map((asset) => (
                  <tr key={asset.id}>
                    <td>
                      <Link to={`/assets/${asset.id}`}>{asset.name}</Link>
                      <span className="tableSub">
                        {asset.code}
                        {asset.type === "vehicle" ? " · Vehicle" : ""}
                      </span>
                    </td>
                    <td>{asset.categoryName}</td>
                    <td>
                      <StatusBadge status={asset.status} />
                    </td>
                    <td>
                      {asset.location || asset.assignedTo || "Not set"}
                      {asset.assignedUserName && (
                        <span className="tableSub">
                          {asset.assignedUserName}
                        </span>
                      )}
                    </td>
                    <td>
                      <WarrantyBadge state={asset.warrantyState} />
                      {asset.warrantyExpiryDate && (
                        <span className="tableSub">
                          {dateOnly(asset.warrantyExpiryDate)}
                        </span>
                      )}
                    </td>
                    <td>{asset.documentCount ?? 0}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </>
  );
}
