import { useRoute } from "preact-iso";

export default function User() {
  const route = useRoute();
  const id = route.params.id;

  return (
    <main class="min-h-screen bg-zinc-950 px-6 py-12 text-zinc-100">
      <section class="mx-auto max-w-3xl rounded-2xl border border-zinc-800 bg-zinc-900 p-8 shadow-xl">
        <p class="mb-3 text-sm font-medium uppercase tracking-wide text-zinc-400">
          Route param
        </p>

        <h1 class="text-4xl font-bold tracking-tight">
          User {id ?? "unknown"}
        </h1>

        <p class="mt-4 text-lg leading-8 text-zinc-300">
          This route is selected by Preact after net/http renders the app shell.
        </p>

        <a class="mt-8 inline-block text-zinc-300 underline" href="/">
          Back home
        </a>
      </section>
    </main>
  );
}
