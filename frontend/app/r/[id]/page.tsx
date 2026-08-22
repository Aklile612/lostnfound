"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { MatchList } from "@/components/MatchList";
import { ReportPayload, getReport, photoUrl, refreshMatches } from "@/lib/api";

export default function ReportPage() {
  const params = useParams<{ id: string }>();
  const [data, setData] = useState<ReportPayload | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!params.id) return;
    getReport(params.id)
      .then(setData)
      .catch((e) => setError(e instanceof Error ? e.message : "Could not load report"));
  }, [params.id]);

  async function onRefresh() {
    if (!params.id) return;
    setBusy(true);
    setError("");
    try {
      setData(await refreshMatches(params.id));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not refresh matches");
    } finally {
      setBusy(false);
    }
  }

  if (error && !data) {
    return <p className="font-mono text-danger">{error}</p>;
  }
  if (!data) {
    return <p className="font-mono text-ink/60">Loading report...</p>;
  }

  const { report, matches } = data;
  return (
    <div className="space-y-8">
      <section className="panel">
        <p className="stamp text-primary">
          {report.type === "lost" ? "Lost item" : "Found item"} · {report.status}
        </p>
        <h1 className="mt-2 text-4xl font-black">{report.title}</h1>
        <p className="mt-2 font-mono text-sm text-ink/70">
          {report.category_label} · {report.location_label}
          {report.location_details ? ` · ${report.location_details}` : ""} · {report.incident_date}
        </p>
        {report.description && <p className="mt-4 text-lg leading-7 text-ink/85">{report.description}</p>}
        {report.unique_features && (
          <p className="mt-2 text-base text-ink/75">
            <span className="font-bold text-ink">Marks:</span> {report.unique_features}
          </p>
        )}
        {report.photos.length > 0 && (
          <div className="mt-4 flex flex-wrap gap-3">
            {report.photos.map((src) => (
              <img
                key={src}
                src={photoUrl(src)}
                alt=""
                className="h-36 w-36 rounded-sm border-[3px] border-ink object-cover"
              />
            ))}
          </div>
        )}
        <button type="button" onClick={onRefresh} disabled={busy} className="btn mt-6 bg-ink text-white shadow-brutal-sm">
          {busy ? "Checking again..." : "Check for new matches"}
        </button>
        {error && <p className="mt-3 font-mono text-sm text-danger">{error}</p>}
      </section>
      <MatchList matches={matches} perspective={report.type} />
    </div>
  );
}
