import { useState, type SubmitEvent } from "react";

type HomeProps = {
  counter?: string;
  user?: string;
};

export default function Home(props: HomeProps) {
  const [user, setUser] = useState(props.user ?? "");

  async function loadUser(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();

    const form = event.currentTarget;
    const name = String(new FormData(form).get("name") ?? "").trim();
    const query = new URLSearchParams({ name });

    try {
      const response = await fetch(`/user?${query}`);

      if (!response.ok) {
        throw new Error(`User island returned ${response.status}`);
      }

      setUser(await response.text());
      window.history.replaceState(null, "", `/?${query}`);
    } catch {
      form.submit();
    }
  }

  return (
    <main className="relative min-h-screen overflow-hidden bg-zinc-950 text-zinc-100">
      <div
        aria-hidden="true"
        className="absolute left-1/2 -top-56 h-128 w-lg -translate-x-1/2 rounded-full bg-indigo-500/15 blur-3xl"
      />

      <div className="relative mx-auto flex min-h-screen max-w-5xl flex-col px-6 py-6 sm:px-8">
        <header className="flex items-center justify-between border-b border-white/10 pb-6">
          <a className="flex items-center gap-3 font-semibold" href="/" aria-label="Zen home">
            <span className="grid h-9 w-9 place-items-center rounded-xl bg-white text-sm font-bold text-zinc-950">
              Z
            </span>
            Zen
          </a>
          <span className="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs text-zinc-400">
            Go + React
          </span>
        </header>

        <section className="grid flex-1 items-center gap-12 py-16 lg:grid-cols-[1.15fr_0.85fr] lg:py-24">
          <div>
            <p className="mb-5 text-sm font-medium text-indigo-300">Small by design.</p>
            <h1 className="max-w-3xl text-4xl font-bold tracking-tight text-balance sm:text-6xl">
              Server-rendered Go apps, without the framework ceremony.
            </h1>
            <p className="mt-6 max-w-2xl text-lg leading-8 text-zinc-400">
              Zen keeps Go, Vite, React, and Tailwind working together while staying out of your way.
            </p>

            <p className="mt-8 text-sm text-zinc-500">Go · Vite · React SSR</p>
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between gap-6 rounded-3xl border border-white/10 bg-white/4 p-6 backdrop-blur sm:p-8">
              <div>
                <p className="text-sm font-medium text-zinc-200">A hydrated island</p>
                <p className="mt-2 text-sm leading-6 text-zinc-500">
                  Server-rendered, then interactive on the client.
                </p>
              </div>
              <div dangerouslySetInnerHTML={{ __html: props.counter ?? "" }} />
            </div>

            <div className="rounded-3xl border border-white/10 bg-white/4 p-6 shadow-2xl shadow-black/20 backdrop-blur sm:p-8">
              <p className="text-sm font-medium text-zinc-200">A normal HTML form</p>
              <p className="mt-2 text-sm leading-6 text-zinc-500">
                Enter a name to render another island on this page.
              </p>

              <form className="mt-6 space-y-4" method="get" action="/" onSubmit={loadUser}>
                <label className="block text-sm text-zinc-300" htmlFor="name">
                  Your name
                </label>
                <input
                  className="w-full rounded-xl border border-white/10 bg-black/20 px-4 py-3 text-zinc-100 outline-none transition placeholder:text-zinc-600 focus:border-indigo-400/70 focus:ring-4 focus:ring-indigo-500/10"
                  id="name"
                  name="name"
                  placeholder="Ada Lovelace"
                  required
                />
                <button
                  className="w-full rounded-xl bg-indigo-400 px-5 py-3 font-semibold text-indigo-950 transition hover:bg-indigo-300"
                  type="submit"
                >
                  Load user
                </button>
              </form>
            </div>

            {user ? <div dangerouslySetInnerHTML={{ __html: user }} /> : null}
          </div>
        </section>

        <footer className="border-t border-white/10 py-6 text-sm text-zinc-600">
          Built with Zen. Make it yours.
        </footer>
      </div>
    </main>
  );
}
