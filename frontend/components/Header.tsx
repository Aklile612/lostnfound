import Link from "next/link";

export function Header() {
  return (
    <header className="border-b-[3px] border-ink bg-surface">
      <div className="mx-auto flex max-w-5xl items-center justify-between gap-4 px-4 py-4 sm:px-6">
        <Link href="/" className="flex items-baseline gap-3">
          <span className="text-3xl font-black leading-none tracking-tight">Reunite</span>
          <span className="stamp hidden text-ink/70 sm:inline">campus lost & found</span>
        </Link>
        <nav className="flex items-center gap-2">
          <Link href="/lost" className="btn bg-surface shadow-brutal-sm">
            Lost
          </Link>
          <Link href="/found" className="btn bg-primary text-white shadow-brutal-sm">
            Found
          </Link>
        </nav>
      </div>
    </header>
  );
}
