type TodoItem = {
  id: number;
  text: string;
};

type TodoProps = {
  todos?: TodoItem[];
};

export default function Todo(props: TodoProps) {
  const todos = props.todos ?? [];

  return (
    <main class="min-h-screen bg-zinc-950 px-6 py-12 text-zinc-100">
      <section class="mx-auto flex w-full max-w-2xl flex-col gap-8 rounded-2xl border border-zinc-800 bg-zinc-900 p-8 shadow-xl">
        <div>
          <p class="mb-3 text-sm font-medium uppercase tracking-wide text-zinc-400">
            Zen todo example
          </p>
          <h1 class="text-4xl font-bold tracking-tight">Todo list</h1>
          <p class="mt-4 text-lg leading-8 text-zinc-300">
            Add a task, remove it, and refresh the page. The list stays in
            memory while the process is running.
          </p>
        </div>

        <form class="flex gap-3" method="post" action="/todos">
          <input
            class="min-w-0 flex-1 rounded-xl border border-zinc-700 bg-zinc-950 px-4 py-3 text-zinc-100 outline-none placeholder:text-zinc-500"
            name="text"
            placeholder="What needs doing?"
            required
          />
          <button
            class="rounded-xl bg-white px-5 py-3 font-semibold text-zinc-950"
            type="submit"
          >
            Add
          </button>
        </form>

        <div class="rounded-2xl border border-zinc-800 bg-zinc-950/60">
          {todos.length === 0 ? (
            <p class="px-4 py-4 text-sm text-zinc-400">No tasks yet.</p>
          ) : (
            <ul class="divide-y divide-zinc-800">
              {todos.map((todo) => (
                <li class="flex items-center justify-between gap-4 px-4 py-4">
                  <span class="min-w-0 break-words text-zinc-100">
                    {todo.text}
                  </span>
                  <form method="post" action={`/todos/${todo.id}/delete`}>
                    <button
                      class="rounded-lg border border-zinc-700 px-3 py-2 text-sm font-medium text-zinc-300"
                      type="submit"
                    >
                      Delete
                    </button>
                  </form>
                </li>
              ))}
            </ul>
          )}
        </div>
      </section>
    </main>
  );
}
