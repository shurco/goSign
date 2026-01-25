export const toDate = (unixTimestamp: number): string =>
  new Date(unixTimestamp * 1000).toLocaleString("en-US", {
    hour12: false,
    timeZoneName: "long"
  });

/** Format a date (ISO string or Date) using pattern: DD, MM, YYYY, MMM, etc. */
export function formatDateByPattern(dateInput: string | Date, format: string): string {
  const date = typeof dateInput === "string" ? new Date(dateInput) : dateInput;
  if (Number.isNaN(date.getTime())) return "";
  const monthFormats: Record<string, "numeric" | "2-digit" | "short" | "long"> = {
    M: "numeric",
    MM: "2-digit",
    MMM: "short",
    MMMM: "long"
  };
  const dayFormats: Record<string, "numeric" | "2-digit"> = {
    D: "numeric",
    DD: "2-digit"
  };
  const yearFormats: Record<string, "numeric" | "2-digit"> = {
    YYYY: "numeric",
    YY: "2-digit"
  };
  const dayMatch = format.match(/D+/);
  const monthMatch = format.match(/M+/);
  const yearMatch = format.match(/Y+/);
  const parts = new Intl.DateTimeFormat([], {
    day: dayMatch ? dayFormats[dayMatch[0]] || "numeric" : "numeric",
    month: monthMatch ? monthFormats[monthMatch[0]] || "numeric" : "numeric",
    year: yearMatch ? yearFormats[yearMatch[0]] || "numeric" : "numeric"
  }).formatToParts(date);
  const dayPart = parts.find((p) => p.type === "day");
  const monthPart = parts.find((p) => p.type === "month");
  const yearPart = parts.find((p) => p.type === "year");
  return format
    .replace(/D+/, dayPart?.value ?? "")
    .replace(/M+/, monthPart?.value ?? "")
    .replace(/Y+/, yearPart?.value ?? "");
}
