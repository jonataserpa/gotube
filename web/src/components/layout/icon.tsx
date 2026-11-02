export type IconName =
  | "menu"
  | "search"
  | "mic"
  | "add"
  | "bell"
  | "home"
  | "bolt"
  | "subscriptions"
  | "library"
  | "history"
  | "videos"
  | "clock"
  | "playlist"
  | "account";

const paths: Record<IconName, string> = {
  menu: "M4 7h16M4 12h16M4 17h16",
  search: "M11 18a7 7 0 1 0 0-14 7 7 0 0 0 0 14ZM20 20l-3.5-3.5",
  mic: "M12 14a3 3 0 0 0 3-3V7a3 3 0 0 0-6 0v4a3 3 0 0 0 3 3ZM8 11a4 4 0 0 0 8 0M12 18v3",
  add: "M12 5v14M5 12h14",
  bell: "M6 16h12l-1.2-2V10a4.8 4.8 0 0 0-9.6 0v4L6 16ZM10 16a2 2 0 0 0 4 0",
  home: "M4 11 12 4l8 7M7 10.5V20h10v-9.5",
  bolt: "M13 3 6 13h5l-1 8 7-10h-5l1-8Z",
  subscriptions: "M4 6h16v12H4V6ZM8 18v2h8v-2",
  library: "M5 5h4v14H5V5ZM11 5h4v14h-4V5ZM17 5h2v14h-2V5Z",
  history: "M5 12a7 7 0 1 0 2-4.9M5 5v4h4M12 8v5l3 2",
  videos: "M4 7h10v10H4V7ZM14 10l6-3v10l-6-3V10Z",
  clock: "M12 7v5l3 2M12 20a8 8 0 1 0 0-16 8 8 0 0 0 0 16Z",
  playlist: "M5 7h14M5 12h14M5 17h9",
  account: "M12 12a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7ZM6 19a6 6 0 0 1 12 0",
};

export function Icon({ name }: { name: IconName }) {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      className="size-5 shrink-0"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d={paths[name]} />
    </svg>
  );
}
