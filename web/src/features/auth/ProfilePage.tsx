import { FormEvent, useState } from "react";
import { PageHeader } from "../../components/PageHeader";
import { useAuth } from "./AuthContext";

export function ProfilePage() {
  const { user, updateProfile } = useAuth();
  const [fullName, setFullName] = useState(user?.fullName ?? "");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [saving, setSaving] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setMessage("");
    setSaving(true);
    try {
      await updateProfile(fullName, password || undefined);
      setPassword("");
      setMessage("Profile updated");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Profile update failed");
    } finally {
      setSaving(false);
    }
  }

  return (
    <>
      <PageHeader title="Profile" eyebrow="Manage your account details" />
      <form className="panel formGrid" onSubmit={submit}>
        <label>
          Email
          <input value={user?.email ?? ""} readOnly />
        </label>
        <label>
          Role
          <input value={user?.role ?? ""} readOnly />
        </label>
        <label>
          Full name
          <input
            required
            value={fullName}
            onChange={(event) => setFullName(event.target.value)}
          />
        </label>
        <label>
          New password
          <input
            autoComplete="new-password"
            minLength={6}
            placeholder="Leave blank to keep current password"
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
          />
        </label>
        {error && <div className="alert wide">{error}</div>}
        {message && <div className="success wide">{message}</div>}
        <div className="formActions wide">
          <button className="primaryButton" disabled={saving} type="submit">
            {saving ? "Saving..." : "Save profile"}
          </button>
        </div>
      </form>
    </>
  );
}
