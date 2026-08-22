"use client";

import { useMemo, useState } from "react";
import { categories, locations } from "@/lib/api";

type Mode = "lost" | "found";

type Props = {
  mode: Mode;
  busy: boolean;
  error: string;
  onSubmit: (form: FormData) => void;
};

export function ReportForm({ mode, busy, error, onSubmit }: Props) {
  const [files, setFiles] = useState<File[]>([]);
  const previews = useMemo(() => files.map((file) => URL.createObjectURL(file)), [files]);
  const today = useMemo(() => {
    const now = new Date();
    const local = new Date(now.getTime() - now.getTimezoneOffset() * 60000);
    return local.toISOString().slice(0, 10);
  }, []);

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    form.delete("photos");
    files.forEach((file) => form.append("photos", file));
    onSubmit(form);
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2">
        <label className="block">
          <span className="label">Category</span>
          <select name="category" required className="field" defaultValue="">
            <option value="" disabled>
              Choose a category
            </option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.label}
              </option>
            ))}
          </select>
        </label>
        <label className="block">
          <span className="label">{mode === "lost" ? "Date lost" : "Date found"}</span>
          <input type="date" name="incident_date" required max={today} defaultValue={today} className="field" />
        </label>
      </div>

      <label className="block">
        <span className="label">{mode === "lost" ? "What did you lose?" : "What did you find?"}</span>
        <input
          name="title"
          className="field"
          placeholder={mode === "lost" ? "Black AirPods case" : "Dark wireless earbud case"}
          required={mode === "lost"}
        />
      </label>

      <label className="block">
        <span className="label">
          {mode === "lost" ? "Description" : "Description (optional if you add a photo)"}
        </span>
        <textarea
          name="description"
          rows={4}
          required={mode === "lost"}
          className="field min-h-28"
          placeholder={
            mode === "lost"
              ? "Color, brand, size, and anything else that identifies it."
              : "Optional. A photo and the place you found it are enough."
          }
        />
      </label>

      <label className="block">
        <span className="label">
          Unique marks {mode === "found" ? "(optional)" : "(optional, but useful)"}
        </span>
        <textarea
          name="unique_features"
          rows={3}
          className="field"
          placeholder="A crack on the left side, a sticker, a broken zipper, engraving..."
        />
      </label>

      <div className="grid gap-4 sm:grid-cols-2">
        <label className="block">
          <span className="label">{mode === "lost" ? "Where did you lose it?" : "Where did you find it?"}</span>
          <select name="location" required className="field" defaultValue="">
            <option value="" disabled>
              Campus place
            </option>
            {locations.map((l) => (
              <option key={l.id} value={l.id}>
                {l.label}
              </option>
            ))}
          </select>
        </label>
        <label className="block">
          <span className="label">More precise spot (optional)</span>
          <input name="location_details" className="field" placeholder="Near the entrance, 2nd floor..." />
        </label>
      </div>

      <div>
        <span className="label">
          Photos {mode === "lost" ? "(optional, helpful as proof)" : "(recommended)"} — up to 3
        </span>
        <input
          type="file"
          accept="image/jpeg,image/png,image/webp"
          multiple
          className="field"
          onChange={(e) => {
            const next = Array.from(e.target.files ?? []).slice(0, 3);
            setFiles(next);
          }}
        />
        {previews.length > 0 && (
          <div className="mt-3 flex gap-3">
            {previews.map((src) => (
              <img key={src} src={src} alt="" className="h-24 w-24 rounded-sm border-[3px] border-ink object-cover" />
            ))}
          </div>
        )}
      </div>

      {mode === "found" && (
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="block">
            <span className="label">Phone number</span>
            <input name="phone" required className="field" placeholder="+251..." />
          </label>
          <label className="block">
            <span className="label">Telegram username</span>
            <input name="telegram" required className="field" placeholder="username" />
          </label>
        </div>
      )}

      {error && (
        <p className="rounded-sm border-[3px] border-danger bg-danger/10 px-4 py-3 font-mono text-sm text-danger">
          {error}
        </p>
      )}

      <button type="submit" disabled={busy} className="btn w-full bg-primary text-white shadow-brutal sm:w-auto">
        {busy
          ? mode === "lost"
            ? "Comparing with found items..."
            : "Saving and checking matches..."
          : mode === "lost"
            ? "Find possible matches"
            : "Report found item"}
      </button>
    </form>
  );
}
