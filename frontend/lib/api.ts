export const categories = [
  { id: "electronics", label: "Electronics" },
  { id: "bags", label: "Bags" },
  { id: "ids", label: "IDs & Documents" },
  { id: "cards", label: "ATM & Bank Cards" },
  { id: "keys", label: "Keys" },
  { id: "clothing", label: "Clothing" },
  { id: "jewelry", label: "Jewelry" },
  { id: "books", label: "Books" },
  { id: "other", label: "Other" },
] as const;

export const locations = [
  { id: "cafeteria", label: "Cafeteria" },
  { id: "library", label: "Library" },
  { id: "coffee_shop", label: "Coffee shop" },
  { id: "lecture_hall", label: "Lecture hall" },
  { id: "dormitory", label: "Dormitory" },
  { id: "gym", label: "Gym" },
  { id: "football_field", label: "Football field" },
  { id: "parking", label: "Parking" },
  { id: "student_center", label: "Student center" },
  { id: "other", label: "Other" },
] as const;

export const apiBase = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export function photoUrl(path: string) {
  if (path.startsWith("http")) return path;
  return `${apiBase}${path}`;
}

export type Report = {
  id: string;
  type: "lost" | "found";
  category: string;
  category_label: string;
  title: string;
  description: string;
  unique_features: string;
  location: string;
  location_label: string;
  location_details: string;
  incident_date: string;
  photos: string[];
  phone?: string;
  telegram?: string;
  status: string;
  created_at: string;
};

export type Match = {
  id: string;
  score: number;
  groq_score: number | null;
  gemini_score: number | null;
  reasoning: string;
  lost_report?: Report;
  found_report?: Report;
};

export type ReportPayload = {
  report: Report;
  matches: Match[];
};

async function readError(res: Response) {
  try {
    const body = await res.json();
    return body.error ?? res.statusText;
  } catch {
    return res.statusText;
  }
}

export async function createReport(kind: "lost" | "found", form: FormData) {
  const res = await fetch(`${apiBase}/api/reports/${kind}`, {
    method: "POST",
    body: form,
  });
  if (!res.ok) throw new Error(await readError(res));
  return (await res.json()) as ReportPayload;
}

export async function getReport(id: string) {
  const res = await fetch(`${apiBase}/api/reports/${id}`, { cache: "no-store" });
  if (!res.ok) throw new Error(await readError(res));
  return (await res.json()) as ReportPayload;
}

export async function refreshMatches(id: string) {
  const res = await fetch(`${apiBase}/api/reports/${id}/matches`, { method: "POST" });
  if (!res.ok) throw new Error(await readError(res));
  return (await res.json()) as ReportPayload;
}

export async function claimFound(id: string) {
  const res = await fetch(`${apiBase}/api/reports/${id}/claim`, { method: "POST" });
  if (!res.ok) throw new Error(await readError(res));
  return (await res.json()) as { report: Report };
}
