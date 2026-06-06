import { useState } from "preact/hooks";

type CounterProps = {
  count: number;
};

export default function Counter(props: CounterProps) {
  const [count, setCount] = useState(props.count);

  return (
    <button
      class="rounded bg-zinc-900 px-3 py-2 text-sm font-medium text-white"
      type="button"
      onClick={() => setCount((current) => current + 1)}
    >
      Count {count}
    </button>
  );
}
