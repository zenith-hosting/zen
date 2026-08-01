import { useState } from "react";

type CounterProps = {
  count: number;
};

export default function Counter(props: CounterProps) {
  const [count, setCount] = useState(props.count);

  return (
    <button
      className="shrink-0 rounded-xl bg-indigo-400 px-4 py-3 text-sm font-semibold text-indigo-950 transition hover:bg-indigo-300"
      type="button"
      onClick={() => setCount((current) => current + 1)}
    >
      Count {count}
    </button>
  );
}
