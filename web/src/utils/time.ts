export function toDate(unixTimestamp: number): string {
  const date = new Date(unixTimestamp * 1000);
  return date.toLocaleString("en-US", {
    hour12: false,
    timeZoneName: "long",
  });
}
