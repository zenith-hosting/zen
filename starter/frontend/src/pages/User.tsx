type UserProps = {
  id: string;
};

export default function User({ id }: UserProps) {
  return (
    <main className="min-h-screen bg-zinc-950 px-6 py-12 text-zinc-100">
      <section className="mx-auto max-w-3xl rounded-2xl border border-zinc-800 bg-zinc-900 p-8 shadow-xl">
        <p className="mb-3 text-sm font-medium uppercase tracking-wide text-zinc-400">
          Route param
        </p>

        <h1 className="text-4xl font-bold tracking-tight">User {id}</h1>

        <p className="mt-4 text-lg leading-8 text-zinc-300">
          This route is selected by React after net/http renders the app shell.
        </p>

        <a className="mt-8 inline-block text-zinc-300 underline" href="/">
          Back home
        </a>
      </section>
    </main>
  );
}
