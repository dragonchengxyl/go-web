export function showAdminToast(
  message: string,
  tone: "default" | "success" | "error" = "default",
) {
  if (typeof document === "undefined") return;

  const colorMap = {
    default: "bg-slate-950 text-white",
    success: "bg-emerald-600 text-white",
    error: "bg-rose-600 text-white",
  } as const;

  const el = document.createElement("div");
  el.className = [
    "fixed right-4 top-4 z-[100] rounded-2xl px-4 py-3 text-sm font-medium shadow-[0_20px_60px_rgba(15,23,42,0.24)]",
    colorMap[tone],
  ].join(" ");
  el.textContent = message;

  document.body.appendChild(el);
  setTimeout(() => el.remove(), 2400);
}
