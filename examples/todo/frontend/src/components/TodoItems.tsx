import { useState } from "preact/hooks";
import { TodoListItem } from "./TodoListItem";
import type { TodoItem, TodoProps } from "../models/todo";

type TodoItemsProps = {
  initialTodos: TodoItem[];
};

export function TodoItems(props: TodoItemsProps) {
  const [todos, setTodos] = useState<TodoItem[]>(props.initialTodos);
  const [error, setError] = useState("");

  async function submitTodo(form: HTMLFormElement) {
    const response = await fetch(form.action, {
      method: "POST",
      body: new FormData(form),
      headers: {
        Accept: "application/json"
      }
    });

    if (!response.ok) {
      throw new Error(`Request failed with ${response.status}`);
    }

    const body = (await response.json()) as TodoProps;
    setTodos(body.todos ?? []);
  }

  async function addTodo(event: SubmitEvent) {
    event.preventDefault();
    setError("");

    const form = event.currentTarget as HTMLFormElement;

    try {
      await submitTodo(form);
      form.reset();
    } catch {
      setError("Could not add that task.");
    }
  }

  async function deleteTodo(event: SubmitEvent) {
    event.preventDefault();
    setError("");

    try {
      await submitTodo(event.currentTarget as HTMLFormElement);
    } catch {
      setError("Could not delete that task.");
    }
  }

  return (
    <>
      <form class="flex gap-3" method="post" action="/todos" onSubmit={addTodo}>
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

      {error ? <p class="text-sm text-red-300">{error}</p> : null}

      <div class="rounded-2xl border border-zinc-800 bg-zinc-950/60">
        {todos.length === 0 ? (
          <p class="px-4 py-4 text-sm text-zinc-400">No tasks yet.</p>
        ) : (
          <ul class="divide-y divide-zinc-800">
            {todos.map((todo) => (
              <TodoListItem todo={todo} onDelete={deleteTodo} key={todo.id} />
            ))}
          </ul>
        )}
      </div>
    </>
  );
}
