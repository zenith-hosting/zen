import Home from "./Home";

type AppProps = {
  url?: string;
  counter?: string;
  user?: string;
};

function NotFound() {
  return (
    <main className="min-h-screen bg-zinc-950 px-6 py-12 text-zinc-100">
      <section className="mx-auto max-w-3xl rounded-2xl border border-zinc-800 bg-zinc-900 p-8 shadow-xl">
        <p className="mb-3 text-sm font-medium uppercase tracking-wide text-zinc-400">
          Not found
        </p>

        <h1 className="text-4xl font-bold tracking-tight">
          This route does not exist.
        </h1>

        <a className="mt-8 inline-block text-zinc-300 underline" href="/">
          Back home
        </a>
      </section>
    </main>
  );
}

export default function App(props: AppProps) {
  const url = new URL(props.url ?? "/", "http://zen.local");

  if (url.pathname === "/") {
    return <Home counter={props.counter} user={props.user} />;
  }

  return <NotFound />;
}
