export function EmptyState({ title, body }: { title: string; body?: string }) {
  return (
    <div className="emptyState">
      <strong>{title}</strong>
      {body && <p>{body}</p>}
    </div>
  );
}
