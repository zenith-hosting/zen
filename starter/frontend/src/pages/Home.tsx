type HomeProps = {
  title?: string;
  body?: string;
};

export default function Home(props: HomeProps) {
  return (
    <main class="min-h-screen bg-zinc-950 px-6 py-12 text-zinc-100">
      <section class="mx-auto max-w-3xl rounded-2xl border border-zinc-800 bg-zinc-900 p-8 shadow-xl">
        <p class="mb-3 text-sm font-medium uppercase tracking-wide text-zinc-400">
          Zen
        </p>

        <h1 class="text-4xl font-bold tracking-tight">
          {props.title ?? "Zen App"}
        </h1>

        <p class="mt-4 text-lg leading-8 text-zinc-300">
          {props.body ?? "Tiny net/http + Vite + Preact SSR glue."}
        </p>

        <form class="mt-8 flex gap-3" method="post" action="/contact">
          <input
            class="min-w-0 flex-1 rounded-xl border border-zinc-700 bg-zinc-950 px-4 py-3 text-zinc-100"
            name="name"
            placeholder="Normal net/http form field"
          />
          <button
            class="rounded-xl bg-white px-5 py-3 font-semibold text-zinc-950"
            type="submit"
          >
            Submit
          </button>
        </form>

        <a class="mt-8 inline-block text-zinc-300 underline" href="/users/42">
          Visit dynamic net/http route
        </a>
      </section>
    </main>
  );
}
