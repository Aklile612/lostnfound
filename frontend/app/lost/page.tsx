"use client";

import Link from "next/link";
import { useState } from "react";
import { MatchList } from "@/components/MatchList";
import { ReportForm } from "@/components/ReportForm";
import { ReportPayload, createReport } from "@/lib/api";

export default function LostPage() {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState<ReportPayload | null>(null);

  async function onSubmit(form: FormData) {
    setBusy(true);
    setError("");
    try {
      const payload = await createReport("lost", form);
      setResult(payload);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not submit the report");
    } finally {
      setBusy(false);
    }
  }

  if (result) {
    return (
      <div className="space-y-6">
        <div>
          <p className="stamp text-primary">Lost report saved</p>
          <h1 className="mt-2 text-4xl font-black">Possible matches</h1>
          <p className="mt-2 text-lg text-ink/80">Keep this link and check again later if nothing looks right now.</p>
          <Link href={`/r/${result.report.id}`} className="mt-3 inline-block font-mono text-sm underline">
            /r/{result.report.id}
          </Link>
        </div>
        <MatchList matches={result.matches} perspective="lost" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <p className="stamp text-primary">No sign-in</p>
        <h1 className="mt-2 text-4xl font-black">I lost something</h1>
        <p className="mt-2 text-lg text-ink/80">
          Describe the item, where you lost it, and any unique marks. A photo helps if you have one.
        </p>
      </div>
      <div className="panel">
        <ReportForm mode="lost" busy={busy} error={error} onSubmit={onSubmit} />
      </div>
    </div>
  );
}
