import { FormEvent } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Archive,
  Download,
  Pencil,
  RotateCcw,
  Trash2,
  Upload,
} from "lucide-react";
import { api, authDownloadHeaders } from "../../api/client";
import type {
  AssetDocument,
  VehicleEmissionRecord,
  VehicleInsuranceRecord,
  VehicleLicenseRecord,
} from "../../api/types";
import { PageHeader } from "../../components/PageHeader";
import { EmptyState } from "../../components/EmptyState";
import { Loading } from "../../components/Loading";
import { StatusBadge, WarrantyBadge } from "../../components/StateBadge";
import { bytes, dateOnly, money } from "./format";

export function AssetDetailPage() {
  const { id } = useParams();
  const assetId = Number(id);
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const {
    data: asset,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["asset", assetId],
    queryFn: () => api.asset(assetId),
    enabled: Number.isFinite(assetId),
  });
  const archive = useMutation({
    mutationFn: () => api.archiveAsset(assetId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["assets"] });
      await queryClient.invalidateQueries({ queryKey: ["asset", assetId] });
    },
  });
  const restore = useMutation({
    mutationFn: () => api.restoreAsset(assetId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["assets"] });
      await queryClient.invalidateQueries({ queryKey: ["asset", assetId] });
    },
  });

  if (isLoading) return <Loading />;
  if (error)
    return (
      <div className="alert">
        {error instanceof Error ? error.message : "Could not load asset"}
      </div>
    );
  if (!asset) return null;

  return (
    <>
      <PageHeader
        title={asset.name}
        eyebrow={`${asset.code} · ${asset.categoryName ?? asset.type}`}
        actions={
          <>
            <Link className="secondaryButton" to={`/assets/${asset.id}/edit`}>
              <Pencil size={16} /> Edit
            </Link>
            {asset.archivedAt ? (
              <button
                className="secondaryButton"
                type="button"
                onClick={() => restore.mutate()}
              >
                <RotateCcw size={16} /> Restore
              </button>
            ) : (
              <button
                className="secondaryButton"
                type="button"
                onClick={() => archive.mutate()}
              >
                <Archive size={16} /> Archive
              </button>
            )}
          </>
        }
      />

      <section className="detailHero">
        <div>
          <span>Status</span>
          <StatusBadge status={asset.status} />
        </div>
        <div>
          <span>Warranty</span>
          <WarrantyBadge state={asset.warrantyState} />
          <strong>
            {dateOnly(asset.warrantyExpiryDate) || "No expiry set"}
          </strong>
        </div>
        <div>
          <span>Documents</span>
          <strong>{asset.documentCount ?? 0}</strong>
        </div>
        <div>
          <span>Purchase price</span>
          <strong>{money(asset.purchasePrice) || "Not set"}</strong>
        </div>
      </section>

      <section className="gridTwo">
        <div className="panel detailList">
          <h2>Asset details</h2>
          <Detail label="Type" value={asset.type} />
          <Detail label="Brand" value={asset.brand} />
          <Detail label="Model" value={asset.model} />
          <Detail label="Serial / VIN" value={asset.serialNumber} />
          <Detail label="Purchase date" value={dateOnly(asset.purchaseDate)} />
          <Detail label="Condition" value={asset.condition} />
          <Detail label="Location" value={asset.location} />
          <Detail
            label="Assigned person"
            value={asset.assignedTo || asset.assignedUserName}
          />
          <Detail label="Notes" value={asset.notes} />
        </div>
        <div className="panel detailList">
          <h2>Warranty</h2>
          <Detail
            label="Start date"
            value={dateOnly(asset.warrantyStartDate)}
          />
          <Detail
            label="Expiry date"
            value={dateOnly(asset.warrantyExpiryDate)}
          />
          <Detail label="State" value={asset.warrantyState} />
          <Detail label="Notes" value={asset.warrantyNotes} />
        </div>
      </section>

      <DocumentSection assetId={asset.id} />
      <ServiceSection assetId={asset.id} />
      {asset.type === "vehicle" && <VehicleSection assetId={asset.id} />}
      <div className="bottomNav">
        <button
          className="secondaryButton"
          type="button"
          onClick={() => navigate("/assets")}
        >
          Back to assets
        </button>
      </div>
    </>
  );
}

function Detail({
  label,
  value,
}: {
  label: string;
  value?: string | number | null;
}) {
  return (
    <div>
      <span>{label}</span>
      <strong>{value || "Not set"}</strong>
    </div>
  );
}

function DocumentSection({ assetId }: { assetId: number }) {
  const queryClient = useQueryClient();
  const { data = [] } = useQuery({
    queryKey: ["documents", assetId],
    queryFn: () => api.documents(assetId),
  });
  const upload = useMutation({
    mutationFn: (form: FormData) => api.uploadDocument(assetId, form),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["documents", assetId] });
      await queryClient.invalidateQueries({ queryKey: ["asset", assetId] });
    },
  });
  const replace = useMutation({
    mutationFn: ({
      documentId,
      form,
    }: {
      documentId: number;
      form: FormData;
    }) => api.replaceDocument(documentId, form),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["documents", assetId] });
      await queryClient.invalidateQueries({ queryKey: ["asset", assetId] });
    },
  });
  const remove = useMutation({
    mutationFn: (documentId: number) => api.deleteDocument(documentId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["documents", assetId] });
      await queryClient.invalidateQueries({ queryKey: ["asset", assetId] });
    },
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    upload.mutate(form);
    event.currentTarget.reset();
  }

  return (
    <section className="panel">
      <h2>Documents</h2>
      <form className="uploadRow" onSubmit={submit}>
        <input name="title" placeholder="Document title" />
        <select name="type" defaultValue="bill_invoice">
          <option value="bill_invoice">Bill / invoice</option>
          <option value="warranty">Warranty document</option>
          <option value="insurance">Insurance document</option>
          <option value="license_registration">License / registration</option>
          <option value="service_receipt">Service receipt</option>
          <option value="manual">Manual</option>
          <option value="other">Other</option>
        </select>
        <input name="notes" placeholder="Notes" />
        <input name="file" type="file" accept=".jpg,.jpeg,.png,.pdf" required />
        <button
          className="primaryButton"
          disabled={upload.isPending}
          type="submit"
        >
          <Upload size={16} /> Upload
        </button>
      </form>
      {upload.error && (
        <div className="alert">
          {upload.error instanceof Error
            ? upload.error.message
            : "Upload failed"}
        </div>
      )}
      {replace.error && (
        <div className="alert">
          {replace.error instanceof Error
            ? replace.error.message
            : "Replace failed"}
        </div>
      )}
      {data.length === 0 ? (
        <EmptyState title="No documents attached" />
      ) : (
        <div className="documentGrid">
          {data.map((doc) => (
            <DocumentCard
              key={doc.id}
              doc={doc}
              onDelete={() => remove.mutate(doc.id)}
              onReplace={(form) => replace.mutate({ documentId: doc.id, form })}
              replacing={replace.isPending}
            />
          ))}
        </div>
      )}
    </section>
  );
}

function DocumentCard({
  doc,
  onDelete,
  onReplace,
  replacing,
}: {
  doc: AssetDocument;
  onDelete(): void;
  onReplace(form: FormData): void;
  replacing: boolean;
}) {
  async function download() {
    const response = await fetch(api.downloadURL(doc.id), {
      headers: authDownloadHeaders(),
    });
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = doc.fileName;
    anchor.click();
    URL.revokeObjectURL(url);
  }

  function replace(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onReplace(new FormData(event.currentTarget));
    event.currentTarget.reset();
  }

  return (
    <article className="documentCard">
      <strong>{doc.title}</strong>
      <span>{doc.type}</span>
      <small>
        {doc.fileName} · {bytes(doc.sizeBytes)}
      </small>
      <div>
        <button
          className="iconButton"
          type="button"
          title="Download"
          onClick={download}
        >
          <Download size={16} />
        </button>
        <button
          className="iconButton danger"
          type="button"
          title="Delete"
          onClick={onDelete}
        >
          <Trash2 size={16} />
        </button>
      </div>
      <form className="replaceRow" onSubmit={replace}>
        <input name="title" type="hidden" defaultValue={doc.title} />
        <input name="type" type="hidden" defaultValue={doc.type} />
        <input name="notes" type="hidden" defaultValue={doc.notes} />
        <input
          aria-label={`Replacement file for ${doc.title}`}
          name="file"
          type="file"
          accept=".jpg,.jpeg,.png,.pdf"
          required
        />
        <button className="secondaryButton" disabled={replacing} type="submit">
          Replace file
        </button>
      </form>
    </article>
  );
}

function ServiceSection({ assetId }: { assetId: number }) {
  const queryClient = useQueryClient();
  const { data = [] } = useQuery({
    queryKey: ["services", assetId],
    queryFn: () => api.services(assetId),
  });
  const mutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) =>
      api.createService(assetId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["services", assetId] });
      await queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      await queryClient.invalidateQueries({ queryKey: ["reminders"] });
    },
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = formValues(event.currentTarget);
    mutation.mutate({
      serviceType: form.serviceType,
      serviceDate: form.serviceDate,
      cost: numberOrUndefined(form.cost),
      vendor: form.vendor,
      description: form.description,
      notes: form.notes,
      mileage: intOrUndefined(form.mileage),
      nextServiceDate: form.nextServiceDate,
      nextServiceMileage: intOrUndefined(form.nextServiceMileage),
    });
    event.currentTarget.reset();
  }

  return (
    <section className="panel">
      <h2>Service and repair history</h2>
      <form className="recordForm" onSubmit={submit}>
        <select name="serviceType" defaultValue="service">
          <option value="service">Service</option>
          <option value="repair">Repair</option>
        </select>
        <input name="serviceDate" type="date" required />
        <input name="vendor" placeholder="Vendor" />
        <input name="cost" type="number" step="0.01" placeholder="Cost" />
        <input name="mileage" type="number" placeholder="Mileage" />
        <input name="nextServiceDate" type="date" />
        <input
          name="nextServiceMileage"
          type="number"
          placeholder="Next mileage"
        />
        <input name="description" placeholder="Description" />
        <input name="notes" placeholder="Notes" />
        <button className="primaryButton" type="submit">
          Add service
        </button>
      </form>
      <RecordTable
        headers={["Date", "Type", "Vendor", "Cost", "Next due", "Description"]}
        rows={data.map((item) => [
          dateOnly(item.serviceDate),
          item.serviceType,
          item.vendor,
          money(item.cost),
          dateOnly(item.nextServiceDate),
          item.description,
        ])}
        empty="No service records"
      />
    </section>
  );
}

function VehicleSection({ assetId }: { assetId: number }) {
  return (
    <>
      <VehicleProfileSection assetId={assetId} />
      <section className="gridTwo">
        <RenewalSection assetId={assetId} kind="insurance" />
        <RenewalSection assetId={assetId} kind="license" />
      </section>
      <section className="gridTwo">
        <RenewalSection assetId={assetId} kind="emission" />
        <FuelSection assetId={assetId} />
      </section>
    </>
  );
}

function VehicleProfileSection({ assetId }: { assetId: number }) {
  const queryClient = useQueryClient();
  const { data } = useQuery({
    queryKey: ["vehicleProfile", assetId],
    queryFn: () => api.vehicleProfile(assetId),
    retry: false,
  });
  const mutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) =>
      api.saveVehicleProfile(assetId, payload),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["vehicleProfile", assetId] }),
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const values = formValues(event.currentTarget);
    mutation.mutate({
      registrationNumber: values.registrationNumber,
      vehicleType: values.vehicleType,
      chassisNumber: values.chassisNumber,
      engineNumber: values.engineNumber,
      odometer: intOrUndefined(values.odometer),
      assignedDriver: values.assignedDriver,
      nextServiceDate: values.nextServiceDate,
      nextServiceMileage: intOrUndefined(values.nextServiceMileage),
      notes: values.notes,
    });
  }

  return (
    <section className="panel">
      <h2>Vehicle profile</h2>
      <form className="formGrid compact" onSubmit={submit}>
        <input
          name="registrationNumber"
          placeholder="Registration number"
          defaultValue={data?.registrationNumber}
          required
        />
        <input
          name="vehicleType"
          placeholder="Vehicle type"
          defaultValue={data?.vehicleType}
        />
        <input
          name="chassisNumber"
          placeholder="Chassis number"
          defaultValue={data?.chassisNumber}
        />
        <input
          name="engineNumber"
          placeholder="Engine number"
          defaultValue={data?.engineNumber}
        />
        <input
          name="odometer"
          type="number"
          placeholder="Odometer"
          defaultValue={data?.odometer}
        />
        <input
          name="assignedDriver"
          placeholder="Assigned driver"
          defaultValue={data?.assignedDriver}
        />
        <input
          name="nextServiceDate"
          type="date"
          defaultValue={dateOnly(data?.nextServiceDate)}
        />
        <input
          name="nextServiceMileage"
          type="number"
          placeholder="Next service mileage"
          defaultValue={data?.nextServiceMileage}
        />
        <input
          className="wide"
          name="notes"
          placeholder="Vehicle notes"
          defaultValue={data?.notes}
        />
        <button className="primaryButton" type="submit">
          Save vehicle profile
        </button>
      </form>
      {mutation.error && (
        <div className="alert">
          {mutation.error instanceof Error
            ? mutation.error.message
            : "Save failed"}
        </div>
      )}
    </section>
  );
}

function RenewalSection({
  assetId,
  kind,
}: {
  assetId: number;
  kind: "insurance" | "license" | "emission";
}) {
  const queryClient = useQueryClient();
  const queryKey = [kind, assetId];
  type RenewalRecord =
    | VehicleInsuranceRecord
    | VehicleLicenseRecord
    | VehicleEmissionRecord;
  const { data = [] } = useQuery<RenewalRecord[]>({
    queryKey,
    queryFn: () => {
      if (kind === "insurance") return api.insurance(assetId);
      if (kind === "license") return api.licenses(assetId);
      return api.emissions(assetId);
    },
  });
  const { data: documents = [] } = useQuery({
    queryKey: ["documents", assetId],
    queryFn: () => api.documents(assetId),
  });
  const mutation = useMutation<RenewalRecord, Error, Record<string, unknown>>({
    mutationFn: (payload: Record<string, unknown>) => {
      if (kind === "insurance") return api.createInsurance(assetId, payload);
      if (kind === "license") return api.createLicense(assetId, payload);
      return api.createEmission(assetId, payload);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey });
      await queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      await queryClient.invalidateQueries({ queryKey: ["reminders"] });
    },
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const values = formValues(event.currentTarget);
    if (kind === "insurance") {
      mutation.mutate({
        provider: values.name,
        policyNumber: values.reference,
        cost: numberOrUndefined(values.cost),
        startDate: values.issueDate,
        expiryDate: values.expiryDate,
        documentId: intOrUndefined(values.documentId),
        notes: values.notes,
      });
    } else if (kind === "license") {
      mutation.mutate({
        renewalType: values.name,
        referenceNumber: values.reference,
        cost: numberOrUndefined(values.cost),
        issueDate: values.issueDate,
        expiryDate: values.expiryDate,
        documentId: intOrUndefined(values.documentId),
        notes: values.notes,
      });
    } else {
      mutation.mutate({
        inspectionType: values.name,
        referenceNumber: values.reference,
        cost: numberOrUndefined(values.cost),
        issueDate: values.issueDate,
        expiryDate: values.expiryDate,
        documentId: intOrUndefined(values.documentId),
        notes: values.notes,
      });
    }
    event.currentTarget.reset();
  }

  return (
    <div className="panel">
      <h2>{kindTitle(kind)}</h2>
      <form className="recordForm single" onSubmit={submit}>
        <input
          name="name"
          placeholder={kind === "insurance" ? "Provider" : "Type"}
        />
        <input name="reference" placeholder="Reference" />
        <input name="cost" type="number" step="0.01" placeholder="Cost" />
        <input name="issueDate" type="date" />
        <input name="expiryDate" type="date" required />
        <select name="documentId" defaultValue="">
          <option value="">No linked document</option>
          {documents.map((doc) => (
            <option key={doc.id} value={doc.id}>
              {doc.title} ({doc.type})
            </option>
          ))}
        </select>
        <input name="notes" placeholder="Notes" />
        <button className="primaryButton" type="submit">
          Add
        </button>
      </form>
      <RecordTable
        headers={["Expiry", "Name", "Reference", "Cost", "Document"]}
        rows={data.map((item) => [
          dateOnly(item.expiryDate),
          "provider" in item
            ? item.provider
            : "renewalType" in item
              ? item.renewalType
              : item.inspectionType,
          "policyNumber" in item ? item.policyNumber : item.referenceNumber,
          money(item.cost),
          documentTitle(documents, item.documentId),
        ])}
        empty={`No ${kind} records`}
      />
    </div>
  );
}

function FuelSection({ assetId }: { assetId: number }) {
  const queryClient = useQueryClient();
  const { data = [] } = useQuery({
    queryKey: ["fuelLogs", assetId],
    queryFn: () => api.fuelLogs(assetId),
  });
  const mutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) =>
      api.createFuelLog(assetId, payload),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["fuelLogs", assetId] }),
  });
  const total = data.reduce((sum, item) => sum + item.cost, 0);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const values = formValues(event.currentTarget);
    mutation.mutate({
      fuelDate: values.fuelDate,
      fuelType: values.fuelType,
      quantity: Number(values.quantity),
      cost: Number(values.cost),
      odometer: intOrUndefined(values.odometer),
      notes: values.notes,
    });
    event.currentTarget.reset();
  }

  return (
    <div className="panel">
      <h2>Fuel logs</h2>
      <p className="muted">Total fuel cost: {money(total)}</p>
      <form className="recordForm single" onSubmit={submit}>
        <input name="fuelDate" type="date" required />
        <input name="fuelType" placeholder="Fuel type" />
        <input
          name="quantity"
          type="number"
          step="0.001"
          placeholder="Quantity"
          required
        />
        <input
          name="cost"
          type="number"
          step="0.01"
          placeholder="Cost"
          required
        />
        <input name="odometer" type="number" placeholder="Odometer" />
        <input name="notes" placeholder="Notes" />
        <button className="primaryButton" type="submit">
          Add fuel
        </button>
      </form>
      <RecordTable
        headers={["Date", "Fuel", "Qty", "Cost", "Odometer"]}
        rows={data.map((item) => [
          dateOnly(item.fuelDate),
          item.fuelType,
          item.quantity.toString(),
          money(item.cost),
          item.odometer?.toString() ?? "",
        ])}
        empty="No fuel logs"
      />
    </div>
  );
}

function RecordTable({
  headers,
  rows,
  empty,
}: {
  headers: string[];
  rows: string[][];
  empty: string;
}) {
  if (rows.length === 0) return <EmptyState title={empty} />;
  return (
    <div className="tableWrap compactTable">
      <table>
        <thead>
          <tr>
            {headers.map((header) => (
              <th key={header}>{header}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, index) => (
            <tr key={index}>
              {row.map((cell, cellIndex) => (
                <td key={cellIndex}>{cell || "Not set"}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function formValues(form: HTMLFormElement) {
  return Object.fromEntries(new FormData(form).entries()) as Record<
    string,
    string
  >;
}

function numberOrUndefined(value: string) {
  return value ? Number(value) : undefined;
}

function intOrUndefined(value: string) {
  return value ? Number.parseInt(value, 10) : undefined;
}

function documentTitle(documents: AssetDocument[], id?: number) {
  if (!id) return "";
  return documents.find((doc) => doc.id === id)?.title ?? `Document #${id}`;
}

function kindTitle(kind: "insurance" | "license" | "emission") {
  if (kind === "insurance") return "Insurance records";
  if (kind === "license") return "License records";
  return "Emission / inspection records";
}
