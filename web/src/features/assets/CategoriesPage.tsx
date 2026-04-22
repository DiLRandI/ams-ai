import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { FormEvent } from "react";
import { api } from "../../api/client";
import { PageHeader } from "../../components/PageHeader";

export function CategoriesPage() {
  return (
    <>
      <PageHeader
        title="Categories"
        eyebrow="Manage asset categories"
      />
      <CategoryManager />
    </>
  );
}

function CategoryManager() {
  const queryClient = useQueryClient();
  const { data: categories = [] } = useQuery({
    queryKey: ["categories"],
    queryFn: api.categories,
  });
  const create = useMutation({
    mutationFn: (payload: { name: string; description: string }) =>
      api.createCategory(payload),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["categories"] }),
  });
  const update = useMutation({
    mutationFn: (payload: { id: number; name: string; description: string }) =>
      api.updateCategory(payload.id, payload),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["categories"] }),
  });

  function createCategory(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    create.mutate({
      name: String(form.get("name") ?? ""),
      description: String(form.get("description") ?? ""),
    });
    event.currentTarget.reset();
  }

  return (
    <section className="panel">
      <h2>Category settings</h2>
      <form className="recordForm single" onSubmit={createCategory}>
        <input name="name" placeholder="New category name" required />
        <input name="description" placeholder="Description" />
        <button className="primaryButton" type="submit">
          Add category
        </button>
      </form>
      <div className="categoryEditor">
        {categories.map((category) => (
          <form
            key={category.id}
            onSubmit={(event) => {
              event.preventDefault();
              const form = new FormData(event.currentTarget);
              update.mutate({
                id: category.id,
                name: String(form.get("name") ?? ""),
                description: String(form.get("description") ?? ""),
              });
            }}
          >
            <input name="name" defaultValue={category.name} />
            <input name="description" defaultValue={category.description} />
            <button className="secondaryButton" type="submit">
              Save
            </button>
          </form>
        ))}
      </div>
    </section>
  );
}
