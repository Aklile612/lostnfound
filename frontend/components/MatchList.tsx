"use client";

import { useState } from "react";
import { Match, Report, claimFound, photoUrl } from "@/lib/api";

function Score({ value }: { value: number }) {
  const rounded = Math.round(value);
  const tone =
    rounded >= 75
      ? "bg-success text-white"
      : rounded >= 55
        ? "bg-primary text-white"
        : "bg-secondary text-ink";
  return (
    <span className={`stamp rounded-sm border-[3px] border-ink px-2.5 py-1 ${tone}`}>{rounded}% match</span>
  );
}

function Place({ report }: { report: Report }) {
  return (
    <p className="font-mono text-sm text-ink/70">
      {report.location_label}
      {report.location_details ? ` · ${report.location_details}` : ""} · {report.incident_date}
    </p>
  );
}

function Photos({ report }: { report: Report }) {
  if (!report.photos?.length) return null;
  return (
    <div className="mt-3 flex gap-2">
      {report.photos.map((src) => (
        <img
          key={src}
          src={photoUrl(src)}
          alt=""
          className="h-28 w-28 rounded-sm border-[3px] border-ink object-cover"
        />
      ))}
    </div>
  );
}

export function MatchList({
  matches,
  perspective,
}: {
  matches: Match[];
  perspective: "lost" | "found";
}) {
  const [claimed, setClaimed] = useState<Record<string, boolean>>({});
  const [err, setErr] = useState("");

  if (matches.length === 0) {
    return (
      <div className="rounded-sm border-[3px] border-dashed border-ink bg-surface p-8 text-center">
        <p className="text-2xl font-black">No likely matches yet</p>
        <p className="mt-3 text-base text-ink/75">
          We only compare items in the same category, still unclaimed, and found within 7 days after the lost date.
          Save your report link and check again later.
        </p>
      </div>
    );
  }

  async function onClaim(id: string) {
    setErr("");
    try {
      await claimFound(id);
      setClaimed((prev) => ({ ...prev, [id]: true }));
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Could not claim this item");
    }
  }

  return (
    <div className="space-y-4">
      {err && (
        <p className="rounded-sm border-[3px] border-danger bg-danger/10 px-4 py-3 font-mono text-sm text-danger">
          {err}
        </p>
      )}
      {matches.map((match) => {
        const other = perspective === "lost" ? match.found_report : match.lost_report;
        if (!other) return null;
        const found = match.found_report;
        return (
          <article key={match.id} className="rounded-sm border-[3px] border-ink bg-surface p-5 shadow-brutal-sm">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h3 className="text-2xl font-black leading-tight">{other.title || other.category_label}</h3>
                <Place report={other} />
              </div>
              <Score value={match.score} />
            </div>
            {other.description && <p className="mt-3 text-base leading-6 text-ink/85">{other.description}</p>}
            {other.unique_features && (
              <p className="mt-2 text-base text-ink/75">
                <span className="font-bold text-ink">Marks:</span> {other.unique_features}
              </p>
            )}
            <Photos report={other} />
            <p className="mt-4 border-l-[3px] border-primary pl-3 text-base leading-6">{match.reasoning}</p>
            {match.groq_score != null || match.gemini_score != null ? (
              <p className="stamp mt-3 text-ink/50">
                {match.groq_score != null ? `Groq ${Math.round(match.groq_score)}` : null}
                {match.groq_score != null && match.gemini_score != null ? " · " : null}
                {match.gemini_score != null ? `Gemini ${Math.round(match.gemini_score)}` : null}
              </p>
            ) : (
              <p className="stamp mt-3 text-ink/50">Local heuristic — add Groq and Gemini keys for model scoring</p>
            )}
            {perspective === "lost" && found && found.status !== "claimed" && !claimed[found.id] && (
              <div className="mt-4 flex flex-wrap items-center gap-3 border-t-[3px] border-ink pt-4">
                <a className="btn bg-ink text-white shadow-brutal-sm" href={`tel:${found.phone}`}>
                  Call {found.phone}
                </a>
                <a
                  className="btn bg-surface shadow-brutal-sm"
                  href={`https://t.me/${found.telegram}`}
                  target="_blank"
                  rel="noreferrer"
                >
                  Telegram @{found.telegram}
                </a>
                <button type="button" onClick={() => onClaim(found.id)} className="font-mono text-sm underline">
                  Mark as claimed
                </button>
              </div>
            )}
            {perspective === "lost" && found && (found.status === "claimed" || claimed[found.id]) && (
              <p className="mt-4 font-mono text-sm font-bold text-success">Marked as claimed.</p>
            )}
          </article>
        );
      })}
    </div>
  );
}
