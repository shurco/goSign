export const toDate = (unixTimestamp: number): string =>
  new Date(unixTimestamp * 1000).toLocaleString("en-US", {
    hour12: false,
    timeZoneName: "long"
  });
