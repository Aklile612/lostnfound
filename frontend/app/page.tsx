import Link from "next/link";

export default function HomePage() {
  return (
    <div className="space-y-8">
      <section className="grid items-stretch gap-6 lg:grid-cols-[1.15fr_0.85fr]">
        <div className="panel relative">
          <p className="stamp text-primary">University lost & found</p>
          <h1 className="mt-4 max-w-xl text-5xl font-black leading-[0.95] tracking-tight sm:text-6xl">
            Lost it on campus. Find who picked it up.
          </h1>
          <p className="mt-6 max-w-lg text-xl leading-7 text-ink/80">
            No account. Report what you lost or found. We drop impossible pairs first, then Groq and Gemini compare
            descriptions, photos, places, dates, and unique marks.
          </p>
        </div>
        <div className="grid gap-4">
          <Link
            href="/lost"
            className="block rounded-sm border-[3px] border-ink bg-primary px-6 py-8 text-white shadow-brutal"
          >
            <p className="stamp text-white/80">I lost something</p>
            <p className="mt-3 text-3xl font-black leading-none">Describe it. See matches.</p>
          </Link>
          <Link
            href="/found"
            className="block rounded-sm border-[3px] border-ink bg-secondary px-6 py-8 text-ink shadow-brutal"
          >
            <p className="stamp">I found something</p>
            <p className="mt-3 text-3xl font-black leading-none">Add a photo and a place.</p>
          </Link>
        </div>
      </section>

      <section className="grid gap-4 md:grid-cols-3">
        {[
          {
            n: "01",
            title: "Hard filter first",
            body: "Same category, still unclaimed, and found within 7 days after it was lost. The models never see impossible rows.",
          },
          {
            n: "02",
            title: "Two models, one score",
            body: "Groq reads the text. Gemini looks at photos when they exist. We combine both, with a local fallback if keys are missing.",
          },
          {
            n: "03",
            title: "Contact the finder",
            body: "Finders leave a phone number and Telegram. Owners see that only after a match clears the threshold.",
          },
        ].map((item) => (
          <article key={item.n} className="rounded-sm border-[3px] border-ink bg-surface p-6 shadow-brutal-sm">
            <p className="stamp text-primary">{item.n}</p>
            <h2 className="mt-3 text-2xl font-black leading-tight">{item.title}</h2>
            <p className="mt-3 text-base leading-6 text-ink/80">{item.body}</p>
          </article>
        ))}
      </section>
    </div>
  );
}
