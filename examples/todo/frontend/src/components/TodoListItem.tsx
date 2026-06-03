import type { TodoItem } from "../models/todo";

type TodoListItemProps = {
  todo: TodoItem;
  onDelete: (event: SubmitEvent) => void;
};

export function TodoListItem(props: TodoListItemProps) {
  return (
    <li class="flex items-center justify-between gap-4 px-4 py-4">
      <span class="min-w-0 break-words text-zinc-100">{props.todo.text}</span>
      <form
        method="post"
        action={`/todos/${props.todo.id}/delete`}
        onSubmit={props.onDelete}
      >
        <button
          class="rounded-lg border border-zinc-700 px-3 py-2 text-sm font-medium text-zinc-300"
          type="submit"
        >
          Delete
        </button>
      </form>
    </li>
  );
}
