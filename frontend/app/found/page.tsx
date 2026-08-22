"use client";

import Link from "next/link";
import { useState } from "react";
import { MatchList } from "@/components/MatchList";
import { ReportForm } from "@/components/ReportForm";
import { ReportPayload, createReport } from "@/lib/api";

export default function FoundPage() {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState<ReportPayload | null>(null);

  async function onSubmit(form: FormData) {
    setBusy(true);
    setError("");
    try {
      const payload = await createReport("found", form);
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
          <p className="stamp text-secondary">Found report saved</p>
          <h1 className="mt-2 text-4xl font-black">Thank you</h1>
          <p className="mt-2 max-w-xl text-lg text-ink/80">
            Your phone and Telegram stay with this report so an owner can reach you if it matches. You do not need an
            account.
          </p>
          <Link href={`/r/${result.report.id}`} className="mt-3 inline-block font-mono text-sm underline">
            View this report
          </Link>
        </div>
        {result.matches.length > 0 && (
          <div className="space-y-3">
            <h2 className="text-2xl font-black">Someone may already be looking for this</h2>
            <MatchList matches={result.matches} perspective="found" />
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <p className="stamp text-secondary">No sign-in</p>
        <h1 className="mt-2 text-4xl font-black">I found something</h1>
        <p className="mt-2 text-lg text-ink/80">
          A photo and the place you found it are enough. Add a short description if you want. We need a phone number
          and Telegram so the owner can contact you.
        </p>
      </div>
      <div className="panel">
        <ReportForm mode="found" busy={busy} error={error} onSubmit={onSubmit} />
      </div>
    </div>
  );
}
