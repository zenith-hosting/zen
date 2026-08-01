type UserProps = {
  name: string;
};

export default function User({ name }: UserProps) {
  return (
    <section className="rounded-3xl border border-indigo-400/20 bg-indigo-400/10 p-6 sm:p-8">
      <p className="text-sm font-medium text-indigo-300">User island</p>
      <p className="mt-2 text-2xl font-semibold text-zinc-100">Hello, {name}.</p>
    </section>
  );
}
