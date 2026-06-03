import { TodoItems } from "../components/TodoItems";
import type { TodoProps } from "../models/todo";

export default function Todo(props: TodoProps) {
  return (
    <main class="min-h-screen bg-zinc-950 px-6 py-12 text-zinc-100">
      <section class="mx-auto flex w-full max-w-2xl flex-col gap-8 rounded-2xl border border-zinc-800 bg-zinc-900 p-8 shadow-xl">
        <div>
          <p class="mb-3 text-sm font-medium uppercase tracking-wide text-zinc-400">
            Zen todo example
          </p>
          <h1 class="text-4xl font-bold tracking-tight">Todo list</h1>
          <p class="mt-4 text-lg leading-8 text-zinc-300">
            Add a task, remove it, and keep moving. The list stays in memory
            while the process is running.
          </p>
        </div>

        <TodoItems initialTodos={props.todos ?? []} />
      </section>
    </main>
  );
}
