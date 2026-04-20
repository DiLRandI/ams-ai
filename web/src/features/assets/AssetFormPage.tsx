import { useEffect } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { type Resolver, useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Link, useNavigate, useParams } from "react-router-dom";
import { z } from "zod";
import { api } from "../../api/client";
import { PageHeader } from "../../components/PageHeader";
import { Loading } from "../../components/Loading";
import { dateOnly } from "./format";

const optionalMoney = z.preprocess(
  (value) => (value === "" || Number.isNaN(value) ? undefined : value),
  z.coerce.number().nonnegative().optional(),
);
const optionalID = z.preprocess(
  (value) => (value === "" || Number.isNaN(value) ? undefined : value),
  z.coerce.number().int().positive().optional(),
);

const schema = z.object({
  type: z.enum(["general", "vehicle"]),
  categoryId: z.coerce.number().min(1, "Category is required"),
  name: z.string().min(1, "Asset name is required"),
  brand: z.string().optional(),
  model: z.string().optional(),
  serialNumber: z.string().optional(),
  purchaseDate: z.string().optional(),
  purchasePrice: optionalMoney,
  status: z.enum(["active", "in_repair", "stored", "retired", "disposed"]),
  condition: z.string().optional(),
  location: z.string().optional(),
  assignedTo: z.string().optional(),
  assignedUserId: optionalID,
  notes: z.string().optional(),
  warrantyStartDate: z.string().optional(),
  warrantyExpiryDate: z.string().optional(),
  warrantyNotes: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

export function AssetFormPage() {
  const params = useParams();
  const assetId = params.id ? Number(params.id) : null;
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { data: categories = [] } = useQuery({
    queryKey: ["categories"],
    queryFn: api.categories,
  });
  const { data: users = [] } = useQuery({
    queryKey: ["users"],
    queryFn: api.users,
  });
  const { data: asset, isLoading } = useQuery({
    queryKey: ["asset", assetId],
    queryFn: () => api.asset(assetId!),
    enabled: Boolean(assetId),
  });

  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema) as unknown as Resolver<FormValues>,
    defaultValues: {
      type: "general",
      status: "active",
      categoryId: 0,
      name: "",
    },
  });
  const assetType = useWatch({ control, name: "type" });

  useEffect(() => {
    if (asset) {
      reset({
        type: asset.type,
        categoryId: asset.categoryId,
        name: asset.name,
        brand: asset.brand,
        model: asset.model,
        serialNumber: asset.serialNumber,
        purchaseDate: dateOnly(asset.purchaseDate),
        purchasePrice: asset.purchasePrice,
        status: asset.status,
        condition: asset.condition,
        location: asset.location,
        assignedTo: asset.assignedTo,
        assignedUserId: asset.assignedUserId,
        notes: asset.notes,
        warrantyStartDate: dateOnly(asset.warrantyStartDate),
        warrantyExpiryDate: dateOnly(asset.warrantyExpiryDate),
        warrantyNotes: asset.warrantyNotes,
      });
    }
  }, [asset, reset]);

  const mutation = useMutation({
    mutationFn: (values: FormValues) => {
      const payload = normalize(values);
      return assetId
        ? api.updateAsset(assetId, payload)
        : api.createAsset(payload);
    },
    onSuccess: async (saved) => {
      await queryClient.invalidateQueries({ queryKey: ["assets"] });
      await queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      navigate(`/assets/${saved.id}`);
    },
  });

  if (assetId && isLoading) return <Loading />;

  return (
    <>
      <PageHeader
        title={assetId ? "Edit asset" : "Add asset"}
        eyebrow={assetId ? asset?.code : "Create an MVP asset record"}
        actions={
          <Link
            className="secondaryButton"
            to={assetId ? `/assets/${assetId}` : "/assets"}
          >
            Cancel
          </Link>
        }
      />
      <form
        className="panel formGrid"
        onSubmit={handleSubmit((values) => mutation.mutate(values))}
      >
        <label>
          Asset type
          <select {...register("type")}>
            <option value="general">General physical asset</option>
            <option value="vehicle">Vehicle asset</option>
          </select>
        </label>
        <label>
          Category
          <select {...register("categoryId")}>
            <option value={0}>Select category</option>
            {categories.map((category) => (
              <option key={category.id} value={category.id}>
                {category.name}
              </option>
            ))}
          </select>
          {errors.categoryId && (
            <span className="fieldError">{errors.categoryId.message}</span>
          )}
        </label>
        <label>
          Asset name
          <input {...register("name")} />
          {errors.name && (
            <span className="fieldError">{errors.name.message}</span>
          )}
        </label>
        <label>
          Status
          <select {...register("status")}>
            <option value="active">Active</option>
            <option value="in_repair">In repair</option>
            <option value="stored">Stored</option>
            <option value="retired">Retired</option>
            <option value="disposed">Disposed</option>
          </select>
        </label>
        <label>
          Brand
          <input {...register("brand")} />
        </label>
        <label>
          Model
          <input {...register("model")} />
        </label>
        <label>
          Serial number / VIN
          <input {...register("serialNumber")} />
        </label>
        <label>
          Condition
          <input {...register("condition")} />
        </label>
        <label>
          Purchase date
          <input type="date" {...register("purchaseDate")} />
        </label>
        <label>
          Purchase price
          <input type="number" step="0.01" {...register("purchasePrice")} />
        </label>
        <label>
          Location
          <input {...register("location")} />
        </label>
        <label>
          Assigned person
          <input {...register("assignedTo")} />
        </label>
        <label>
          Assigned app user
          <select {...register("assignedUserId")}>
            <option value="">None</option>
            {users.map((user) => (
              <option key={user.id} value={user.id}>
                {user.fullName}
              </option>
            ))}
          </select>
        </label>
        <label>
          Warranty start
          <input type="date" {...register("warrantyStartDate")} />
        </label>
        <label>
          Warranty expiry
          <input type="date" {...register("warrantyExpiryDate")} />
        </label>
        <label className="wide">
          Warranty notes
          <textarea rows={3} {...register("warrantyNotes")} />
        </label>
        <label className="wide">
          Notes
          <textarea rows={4} {...register("notes")} />
        </label>
        {assetType === "vehicle" && (
          <div className="wide callout">
            Vehicle-specific registration, insurance, license, emission,
            service, and fuel details are managed on the asset detail page after
            saving.
          </div>
        )}
        {mutation.error && (
          <div className="alert wide">
            {mutation.error instanceof Error
              ? mutation.error.message
              : "Save failed"}
          </div>
        )}
        <div className="formActions wide">
          <button
            className="primaryButton"
            disabled={mutation.isPending}
            type="submit"
          >
            {mutation.isPending ? "Saving..." : "Save asset"}
          </button>
        </div>
      </form>
    </>
  );
}

function normalize(values: FormValues) {
  return {
    ...values,
    brand: values.brand ?? "",
    model: values.model ?? "",
    serialNumber: values.serialNumber ?? "",
    purchaseDate: values.purchaseDate || "",
    purchasePrice: values.purchasePrice,
    condition: values.condition ?? "",
    location: values.location ?? "",
    assignedTo: values.assignedTo ?? "",
    assignedUserId: values.assignedUserId,
    notes: values.notes ?? "",
    warrantyStartDate: values.warrantyStartDate || "",
    warrantyExpiryDate: values.warrantyExpiryDate || "",
    warrantyNotes: values.warrantyNotes ?? "",
  };
}
